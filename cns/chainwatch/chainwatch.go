package chainwatch

import (
	"time"

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

		ns, err := q.ChainsByName(chains...)
		if err != nil {
			i.l.Errorw("cannot get chains from k8s", "error", err)
			continue
		}

		for _, n := range ns {
			if n.Status.Phase != v1.PhaseRunning {
				i.l.Debugw("chain not in running phase", "name", n.Name, "phase", n.Status.Phase)
				continue
			}

			i.l.Debugw("chain finished running", "name", n.Name)
			if err := i.c.RemoveChain(n.Name); err != nil {
				i.l.Errorw("cannot delete chain name from cache", "error", err)
				continue
			}

			// TODO: fetch denoms and write them into database
		}
	}
}
