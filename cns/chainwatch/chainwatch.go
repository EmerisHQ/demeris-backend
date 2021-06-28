package chainwatch

import (
	"errors"
	"fmt"
	"time"

	"github.com/allinbits/demeris-backend/models"

	"github.com/allinbits/demeris-backend/cns/database"

	"github.com/allinbits/demeris-backend/utils/k8s/operator"

	v1 "github.com/allinbits/starport-operator/api/v1"

	"github.com/allinbits/demeris-backend/utils/k8s"

	"go.uber.org/zap"
	kube "sigs.k8s.io/controller-runtime/pkg/client"
)

//go:generate stringer -type=chainStatus
type chainStatus uint

const (
	starting chainStatus = iota
	running
	relayerConnecting
	done
)

var maxGas int64 = 6000000

type Instance struct {
	l                *zap.SugaredLogger
	k                kube.Client
	defaultNamespace string
	c                *Connection
	db               *database.Instance
	statusMap        map[string]chainStatus
}

func New(
	l *zap.SugaredLogger,
	k kube.Client,
	defaultNamespace string,
	c *Connection,
	db *database.Instance,
) *Instance {
	return &Instance{
		l:                l,
		k:                k,
		defaultNamespace: defaultNamespace,
		c:                c,
		db:               db,
		statusMap:        map[string]chainStatus{},
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
			chainStatus, found := i.statusMap[chain.Name]
			if !found {
				chainStatus = starting
				i.statusMap[chain.Name] = chainStatus
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
					i.statusMap[chain.Name] = starting
					continue
				}

				// chain is now in running phase
				i.statusMap[chain.Name] = running
				i.l.Debugw("chain status update", "name", chain.Name, "new_status", running.String())
			case running:
				if err := i.chainFinished(chains[idx]); err != nil {
					i.l.Errorw("cannot execute chain finished routine", "error", err)
					continue
				}

				i.statusMap[chain.Name] = relayerConnecting
			case relayerConnecting:
				// TODO: query channels from db if any

				relayer, err := q.Relayer()
				if err != nil {
					i.l.Errorw("cannot get relayer", "error", err)
					continue
				}

				phase := relayer.Status.Phase
				if phase != v1.RelayerPhaseRunning {
					amt, err := i.db.ChainAmount()
					if err != nil {
						i.l.Errorw("cannot get amount of chains", "error", err)
						continue
					}

					if amt == 1 {
						i.statusMap[chain.Name] = done
					}
					continue
				}

				if err := i.relayerFinished(chain, relayer); err != nil {
					i.l.Debugw("error while running relayerfinished", "error", err)
					continue
				}

				i.statusMap[chain.Name] = done
			case done:
				if err := i.c.RemoveChain(chain); err != nil {
					i.l.Errorw("cannot remove chain from redis", "error", err)
				}

				delete(i.statusMap, chain.Name)
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

	relayerConfig.GasPrice = v1.GasPriceConfig{}
	relayerConfig.MaxGas = &maxGas

	relayer.Spec.Chains = append(relayer.Spec.Chains, &relayerConfig)

	ctc, alreadyConnected, err := i.chainsConnectedAndToConnectTo(cfg.NodesetName)
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
		relayer.Spec.Paths = append(relayer.Spec.Paths, v1.PathConfig{
			SideA: cfg.NodesetName,
			SideB: ccc,
		})
	}

	var execErr error
	if errors.Is(err, k8s.ErrNotFound) || relayerWasEmpty {
		relayer.Namespace = i.defaultNamespace
		relayer.Name = "relayer"
		relayer.ObjectMeta.Name = "relayer"
		execErr = q.AddRelayer(relayer)
	} else {
		i.l.Debugw("relayer configuration existing", "configuration", relayer)
		execErr = q.UpdateRelayer(relayer)
	}

	if execErr != nil {
		return fmt.Errorf("cannot update or add relayer, %w", execErr)
	}

	return nil
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

func (i *Instance) chainsConnectedAndToConnectTo(chainName string) ([]string, []connectedChain, error) {
	chains, err := i.db.Chains()
	if err != nil {
		return nil, nil, err
	}

	var ret []string
	var connected []connectedChain

	for _, c := range chains {
		if c.ChainName == chainName {
			continue
		}

		conns, err := i.db.ChannelsBetweenChains(chainName, c.ChainName)
		if err != nil {
			return nil, nil, fmt.Errorf("cannot scan channels between chain %s and %s, %w", chainName, c.ChainName, err)
		}

		i.l.Debugw("chains between", "chainName", chainName, "other", c.ChainName, "conns", conns)

		if conns == nil || len(conns) == 0 {
			ret = append(ret, c.ChainName) // c.ChainName is not connected to chainName
		} else {
			cc := connectedChain{
				chainName:        chainName,
				counterChainName: c.ChainName,
			}

			for chanID, counterChanID := range conns {
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
	chains, err := i.db.Chains()
	if err != nil {
		return err
	}

	chainsMap := map[string]models.Chain{}
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

		if err := i.db.AddChain(chain); err != nil {
			return fmt.Errorf("error while updating chain %s, %w", chain.ChainName, err)
		}

		if err := i.db.AddChain(counterChain); err != nil {
			return fmt.Errorf("error while updating chain %s, %w", counterChain.ChainName, err)
		}
	}

	return nil
}

func (i *Instance) setPrimaryChannel(_ Chain, relayer v1.Relayer) error {
	chainsMap := map[string]models.Chain{}

	chains, err := i.db.Chains()
	if err != nil {
		return err
	}

	for _, chain := range chains {
		i.l.Debugw("chain read", "chainID", chain.NodeInfo.ChainID, "name", chain.ChainName)
		chainsMap[chain.NodeInfo.ChainID] = chain
	}

	paths := relayer.Status.Paths
	for chainID, chain := range chainsMap {
		i.l.Debugw("iterating chainsmap", "chainID", chainID)

		for _, path := range paths {
			i.l.Debugw("iterating path", "path", path)
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
				chain.PrimaryChannel[counterparty.ChainName] = value.ChannelID
			}
		}

		i.l.Debugw("new primary channel struct", "data", chain.PrimaryChannel)

		if err := i.db.AddChain(chain); err != nil {
			return fmt.Errorf("error while updating chain %s, %w", chain.ChainName, err)
		}
	}

	return nil
}
