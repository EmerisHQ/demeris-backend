package chainwatch

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/allinbits/demeris-backend/cns/cnsdb"
	"github.com/allinbits/demeris-backend/models"
	"github.com/allinbits/demeris-backend/utils"

	"github.com/allinbits/demeris-backend/utils/k8s/operator"

	v1 "github.com/allinbits/starport-operator/api/v1"

	"github.com/allinbits/demeris-backend/utils/k8s"

	"go.uber.org/zap"
	kube "sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	relayerDebugLogLevel       = "debug"
	maxGas               int64 = 6000000
	clearPacketsInterval int64 = 600
)

type Instance struct {
	l                *zap.SugaredLogger
	k                kube.Client
	defaultNamespace string
	c                *Connection
	db               *cnsdb.Queries
	relayerDebug     bool
}

func New(
	l *zap.SugaredLogger,
	k kube.Client,
	defaultNamespace string,
	c *Connection,
	db *cnsdb.Queries,
	relayerDebug bool,
) *Instance {
	return &Instance{
		l:                l,
		k:                k,
		defaultNamespace: defaultNamespace,
		c:                c,
		db:               db,
		relayerDebug:     relayerDebug,
	}

}

func (i *Instance) Run() {
	for range time.Tick(1 * time.Second) {
		chains, err := i.c.Chains()
		if err != nil {
			i.l.Errorw("cannot get chains from redis", "error", err)
			continue
		}

		if chains == nil {
			continue
		}

		i.l.Debugw("chains in cache", "list", chains)

		for idx, chain := range chains {
			chainStatus, found, err := i.c.ChainStatus(chain.Name)
			if err != nil {
				i.l.Errorw("cannot query chain status from redis at beginning of chains loop", "chainName", chain.Name, "error", err)
				continue
			}

			if !found {
				chainStatus = starting
				if err := i.c.SetChainStatus(chain.Name, chainStatus); err != nil {
					i.l.Errorw("cannot set new chain status in redis", "chainName", chain.Name, "error", err, "newStatus", starting.String())
					continue
				}
			}

			q := k8s.Querier{
				Client:    i.k,
				Namespace: i.defaultNamespace,
			}

			n, err := q.ChainByName(chain.Name)
			if err != nil {
				i.l.Errorw("cannot get chains from k8s", "error", err)
				continue
			}

			i.l.Debugw("chain status", "name", chain.Name, "status", chainStatus.String())

			switch chainStatus {
			case starting:
				if n.Status.Phase != v1.PhaseRunning {
					i.l.Debugw("chain not in running phase", "name", n.Name, "phase", n.Status.Phase)
					if err := i.c.SetChainStatus(chain.Name, starting); err != nil {
						i.l.Errorw("cannot set chain status in redis", "chainName", chain.Name, "error", err, "newStatus", starting.String())
						continue
					}
					continue
				}

				// chain is now in running phase
				if err := i.c.SetChainStatus(chain.Name, running); err != nil {
					i.l.Errorw("cannot set chain status in redis", "chainName", chain.Name, "error", err, "newStatus", running.String())
					continue
				}

				i.l.Debugw("chain status update", "name", chain.Name, "new_status", running.String())
			case running:
				if err := i.chainFinished(chains[idx]); err != nil {
					i.l.Errorw("cannot execute chain finished routine", "error", err)
					continue
				}

				if err := i.c.SetChainStatus(chain.Name, relayerConnecting); err != nil {
					i.l.Errorw("cannot set chain status in redis", "chainName", chain.Name, "error", err, "newStatus", relayerConnecting.String())
					continue
				}
			case relayerConnecting:
				relayer, err := q.Relayer()
				if err != nil {
					i.l.Errorw("cannot get relayer", "error", err)
					continue
				}

				amt, err := i.db.ChainAmount(context.Background())
				if err != nil {
					i.l.Errorw("cannot get amount of chains", "error", err)
					continue
				}

				chainStatuses := relayer.Status.ChainStatuses

				if amt != int64(len(chainStatuses)) {
					continue // corner case where the chain gets added, previous chains are already connected, but the operator still reports "Running" because the
					// reconcile cycle didn't get up yet.
				}

				phase := relayer.Status.Phase
				if phase != v1.RelayerPhaseRunning {
					if amt == 1 {
						if err := i.c.SetChainStatus(chain.Name, done); err != nil {
							i.l.Errorw("cannot set chain status in redis", "chainName", chain.Name, "error", err, "newStatus", done.String())
							continue
						}
					}
					continue
				}

				if err := i.relayerFinished(chain, relayer); err != nil {
					i.l.Debugw("error while running relayerfinished", "error", err)
					continue
				}

				if err := i.c.SetChainStatus(chain.Name, done); err != nil {
					i.l.Errorw("cannot set chain status in redis", "chainName", chain.Name, "error", err, "newStatus", done.String())
					continue
				}
			case done:
				if err := i.c.RemoveChain(chain); err != nil {
					i.l.Errorw("cannot remove chain from redis", "error", err)
				}

				if err := i.c.DeleteChainStatus(chain.Name); err != nil {
					i.l.Errorw("cannot delete chain status in redis", "chainName", chain.Name, "error", err)
					continue
				}
			}

		}
	}
}

func (i *Instance) chainFinished(chain Chain) error {
	// create relayers
	if err := i.createRelayer(chain); err != nil {
		return err
	}
	return nil
}

func (i *Instance) createRelayer(chain Chain) error {
	relayerDenom, err := i.relayerDenom(chain.Name)
	if err != nil {
		return err
	}

	q := k8s.Querier{
		Client:    i.k,
		Namespace: i.defaultNamespace,
	}

	cfg := operator.RelayerConfig{
		NodesetName:   chain.Name,
		HDPath:        chain.HDPath,
		AccountPrefix: chain.AddressPrefix,
	}

	if chain.HasFaucet {
		cfg.FaucetName = fmt.Sprintf("%s-faucet", cfg.NodesetName)
	}

	relayerConfig, err := operator.BuildRelayer(cfg)
	if err != nil {
		return err
	}

	relayer, err := q.Relayer()
	if err != nil && !errors.Is(err, k8s.ErrNotFound) {
		return fmt.Errorf("cannot query relayer, %w", err)
	}

	relayerWasEmpty := len(relayer.Spec.Chains) == 0

	for _, existingChain := range relayer.Spec.Chains {
		if existingChain.Nodeset == relayerConfig.Nodeset {
			return nil // existing chain, somehow...
		}
	}

	gasPrice := fmt.Sprintf("%.2f", relayerDenom.GasPriceLevels.Average)

	relayerConfig.GasPrice = v1.GasPriceConfig{
		Price: &gasPrice,
		Denom: &relayerDenom.Name,
	}
	relayerConfig.MaxGas = &maxGas

	relayerConfig.MaxMsgNum = &chain.RelayerConfiguration.MaxMsgNum
	relayerConfig.MaxGas = &chain.RelayerConfiguration.MaxGas
	relayerConfig.ClockDrift = &chain.RelayerConfiguration.ClockDrift
	relayerConfig.TrustingPeriod = &chain.RelayerConfiguration.TrustingPeriod

	relayer.Spec.Chains = append(relayer.Spec.Chains, &relayerConfig)

	if !chain.SkipChannelCreation {
		// if we don't want to skip channel creation, always create new paths
		ctc, alreadyConnected, err := i.chainsConnectedAndToConnectTo(cfg.NodesetName, true)
		if err != nil {
			i.l.Debugw("cannot query chains to connect to", "error", err)
			return err
		}

		i.l.Debugw("chains to connect to", "ctc", ctc)
		i.l.Debugw("chains already connected", "chains", alreadyConnected, "how many", len(alreadyConnected))

		if len(alreadyConnected) != 0 {
			if err := i.updateAlreadyConnected(alreadyConnected); err != nil {
				i.l.Debugw("cannot update already connected chains", "error", err)
				return err
			}
		}

		for _, ccc := range ctc {
			newPath := v1.PathConfig{
				SideA: cfg.NodesetName,
				SideB: ccc,
			}

			if pathDuplicate(newPath, relayer.Spec.Paths) {
				continue
			}

			relayer.Spec.Paths = append(relayer.Spec.Paths, newPath)
		}
	} else {
		i.l.Debugw("skipping channel creation", "chainName", cfg.NodesetName)
	}

	var execErr error
	if errors.Is(err, k8s.ErrNotFound) || relayerWasEmpty {
		i.l.Debugw("creating new relayer instance", "debugMode", i.relayerDebug)
		relayer.Namespace = i.defaultNamespace
		relayer.Name = "relayer"
		relayer.ObjectMeta.Name = "relayer"
		relayer.Spec.Filter = true
		relayer.Spec.ClearPacketsInterval = &clearPacketsInterval

		if i.relayerDebug {
			relayer.Spec.LogLevel = &relayerDebugLogLevel
		}

		execErr = q.AddRelayer(relayer)
	} else {
		i.l.Debugw("relayer configuration existing", "configuration", relayer)

		if i.relayerDebug {
			relayer.Spec.LogLevel = &relayerDebugLogLevel
		}

		execErr = q.UpdateRelayer(relayer)
	}

	if execErr != nil {
		return fmt.Errorf("cannot update or add relayer, %w", execErr)
	}

	return nil
}

func pathDuplicate(newConfig v1.PathConfig, paths []v1.PathConfig) bool {
	flipped := v1.PathConfig{
		SideA: newConfig.SideB,
		SideB: newConfig.SideA,
	}

	for _, path := range paths {
		if path == newConfig || path == flipped {
			return true
		}
	}

	return false
}

func (i *Instance) relayerFinished(chain Chain, relayer v1.Relayer) error {
	if err := i.setPrimaryChannel(chain, relayer); err != nil {
		return err
	}

	return nil
}

type connectedChain struct {
	chainName        string
	counterChainName string
	channel          string
	counterChannel   string
}

func (c connectedChain) String() string {
	return fmt.Sprintf("%s, %s, %s, %s", c.chainName, c.counterChainName, c.channel, c.counterChannel)
}

func (i *Instance) chainsConnectedAndToConnectTo(chainName string, alwaysConnect bool) ([]string, []connectedChain, error) {
	sourceChain, err := i.db.Chain(context.Background(), chainName)
	if err != nil {
		return nil, nil, err
	}

	chains, err := i.db.Chains(context.Background())
	if err != nil {
		return nil, nil, err
	}

	var ret []string
	var connected []connectedChain

	for _, c := range chains {
		if c.ChainName == chainName {
			continue
		}

		i.l.Debugw("querying channels between chains", "chainName", chainName, "destination", c.ChainName, "chainID", sourceChain.NodeInfo.ChainID)
		conns, err := i.db.ChannelsBetweenChains(chainName, c.ChainName, sourceChain.NodeInfo.ChainID)
		if err != nil {
			return nil, nil, fmt.Errorf("cannot scan channels between chain %s and %s, %w", chainName, c.ChainName, err)
		}

		i.l.Debugw("chains between", "chainName", chainName, "other", c.ChainName, "conns", conns)

		if conns == nil || len(conns) == 0 || alwaysConnect {
			ret = append(ret, c.ChainName) // c.ChainName is not connected to chainName, or alwaysConnect is true
		} else {
			cc := connectedChain{
				chainName:        chainName,
				counterChainName: c.ChainName,
			}

			for counterChanID, chanID := range conns {
				cc.channel = chanID
				cc.counterChannel = counterChanID
				break
			}

			i.l.Debugw("new connected chain", "chain", cc)

			connected = append(connected, cc) // c.ChainName is connected via cc.counterChannel, chainName is connected via cc.channel
		}
	}

	i.l.Debugw("returning data from chains connected func", "ret", ret, "connected", connected)

	return ret, connected, nil
}

func (i *Instance) updateAlreadyConnected(connected []connectedChain) error {
	chains, err := i.db.Chains(context.Background())
	if err != nil {
		return err
	}

	chainsMap := map[string]cnsdb.Chain{}
	for _, c := range chains {
		chainsMap[c.ChainName] = c
	}

	for _, c := range connected {
		chain, ok := chainsMap[c.chainName]
		if !ok {
			return fmt.Errorf("found chain %s in connectedChain but not into chains database", c.chainName)
		}

		counterChain, ok := chainsMap[c.counterChainName]
		if !ok {
			return fmt.Errorf("found counterparty chain %s in connectedChain but not into chains database", c.counterChainName)
		}

		chain.PrimaryChannel[c.counterChainName] = c.counterChannel
		counterChain.PrimaryChannel[c.chainName] = c.channel

		if err := i.db.AddChain(context.Background(), utils.GetAddChainParams(chain)); err != nil {
			return fmt.Errorf("error while updating chain %s, %w", chain.ChainName, err)
		}

		if err := i.db.AddChain(context.Background(), utils.GetAddChainParams(counterChain)); err != nil {
			return fmt.Errorf("error while updating chain %s, %w", counterChain.ChainName, err)
		}
	}

	return nil
}

func (i *Instance) setPrimaryChannel(_ Chain, relayer v1.Relayer) error {
	chainsMap := map[string]cnsdb.Chain{}

	chains, err := i.db.Chains(context.Background())
	if err != nil {
		return err
	}

	for _, chain := range chains {
		i.l.Debugw("chain read", "chainID", chain.NodeInfo.ChainID, "name", chain.ChainName)
		chainsMap[chain.NodeInfo.ChainID] = chain
	}

	result := i.updatePrimaryChannelForChain(chainsMap, relayer)

	for _, chain := range result {
		if err := i.db.AddChain(context.Background(), utils.GetAddChainParams(chain)); err != nil {
			return fmt.Errorf("error while updating chain %s, %w", chain.ChainName, err)
		}
	}

	return nil
}

func (i *Instance) updatePrimaryChannelForChain(chainsMap map[string]cnsdb.Chain, relayer v1.Relayer) map[string]cnsdb.Chain {

	paths := relayer.Status.Paths
	for chainID, chain := range chainsMap {
		i.l.Debugw("iterating chainsmap", "chainID", chainID)

		for _, path := range paths {
			i.l.Debugw("iterating path", "path", path)

			if _, found := path[chainID]; !found {
				i.l.Debugw("skipping path since it's not related to me", "chain", chain.ChainName)
				continue
			}

			for counterpartyChainID, value := range path {
				i.l.Debugw("beginning of path iteration", "counterpartyChainID", counterpartyChainID, "chainID", chainID)
				if counterpartyChainID == chainID {
					i.l.Debugw("found ourselves", "chainID", chainID)
					continue
				}

				counterparty, found := chainsMap[counterpartyChainID]
				i.l.Debugw("found counterparty in chainsMap", "counterparty name", counterparty.ChainName, "found", found)

				if !found {
					i.l.Panicw("found counterparty chain which isn't in chainsMap", "chainsMap", chainsMap, "counterparty", counterpartyChainID)
				}

				i.l.Debugw("updating chain", "chain to be update", chainsMap[chainID].ChainName, "counterparty", counterparty.ChainName, "value", value.ChannelID)
				if _, ok := chain.PrimaryChannel[counterparty.ChainName]; ok {
					// don't overwrite a primary channel that was set before
					continue
				}

				chain.PrimaryChannel[counterparty.ChainName] = path[chainID].ChannelID
			}
		}

		i.l.Debugw("new primary channel struct", "data", chain.PrimaryChannel)

		chainsMap[chainID] = chain
	}

	return chainsMap
}

func (i *Instance) relayerDenom(chainName string) (models.Denom, error) {
	chain, err := i.db.Chain(context.Background(), chainName)
	if err != nil {
		return models.Denom{}, err
	}

	return chain.RelayerToken(), nil
}
