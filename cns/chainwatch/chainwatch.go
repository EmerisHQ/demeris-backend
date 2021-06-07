package chainwatch

import (
	"errors"
	"fmt"
	"time"

	"github.com/allinbits/demeris-backend/utils/k8s/operator"

	v1 "github.com/allinbits/starport-operator/api/v1"

	"github.com/allinbits/demeris-backend/utils/k8s"

	"go.uber.org/zap"
	kube "sigs.k8s.io/controller-runtime/pkg/client"
)

type Instance struct {
	l *zap.SugaredLogger
	k kube.Client
	c *Connection
}

func New(
	l *zap.SugaredLogger,
	k kube.Client,
	c *Connection,
) *Instance {
	return &Instance{
		l: l,
		k: k,
		c: c,
	}

}

func (i *Instance) Run() {
	for range time.Tick(1 * time.Second) {
		chains, err := i.c.Chains()
		if err != nil {
			i.l.Errorw("cannot get chains from redis", "error", err)
			continue
		}

		i.l.Debugw("chains in cache", "list", chains)

		q := k8s.Querier{Client: i.k}

		var chainsNames []string
		for _, cc := range chains {
			chainsNames = append(chainsNames, cc.Name)
		}

		ns, err := q.ChainsByName(chainsNames...)
		if err != nil {
			i.l.Errorw("cannot get chains from k8s", "error", err)
			continue
		}

		for idx, n := range ns {
			if n.Status.Phase != v1.PhaseRunning {
				i.l.Debugw("chain not in running phase", "name", n.Name, "phase", n.Status.Phase)
				continue
			}

			if err := i.chainFinished(chains[idx]); err != nil {
				i.l.Errorw("cannot execute chain finished routine", "error", err)
			}
		}
	}
}

func (i *Instance) chainFinished(chain Chain) error {
	// create relayers
	if err := i.createRelayer(chain); err != nil {
		return err
	}

	i.l.Debugw("chain finished running", "name", chain.Name)
	if err := i.c.RemoveChain(chain); err != nil {
		return err
	}

	return nil
}

func (i *Instance) createRelayer(chain Chain) error {
	q := k8s.Querier{Client: i.k}

	cfg := operator.RelayerConfig{
		NodesetName:   chain.Name,
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
		return err
	}

	relayer.Spec.Chains = append(relayer.Spec.Chains, &relayerConfig)

	var execErr error
	if errors.Is(err, k8s.ErrNotFound) {
		relayer.Namespace = "default"
		relayer.ObjectMeta.Name = "relayer"
		execErr = q.AddRelayer(relayer)
	} else {
		execErr = q.UpdateRelayer(relayer)
	}

	return execErr
}
