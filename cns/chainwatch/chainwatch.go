package chainwatch

import (
	"errors"
	"fmt"
	"time"

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

type Instance struct {
	l         *zap.SugaredLogger
	k         kube.Client
	c         *Connection
	db        *database.Instance
	statusMap map[string]chainStatus
}

func New(
	l *zap.SugaredLogger,
	k kube.Client,
	c *Connection,
	db *database.Instance,
) *Instance {
	return &Instance{
		l:         l,
		k:         k,
		c:         c,
		db:        db,
		statusMap: map[string]chainStatus{},
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

			q := k8s.Querier{Client: i.k}

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
					if len(chains) == 1 {
						i.statusMap[chain.Name] = done
					}
					continue
				}

				// TODO: write primary channels to db

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
		i.l.Debugw("relayer configuration existing", "configuration", relayer)
		execErr = q.UpdateRelayer(relayer)
	}

	return execErr
}
