package k8s

import (
	"context"
	"fmt"

	v1 "github.com/allinbits/starport-operator/api/v1"
	kube "sigs.k8s.io/controller-runtime/pkg/client"
)

type Querier struct {
	Client kube.Client
}

func (q Querier) ChainRunning(name string) (bool, error) {
	var chainList v1.NodeSetList

	if err := q.Client.List(context.TODO(), &chainList, kube.MatchingFields{
		"metadata.name": name,
	}); err != nil {
		return false, err
	}

	if len(chainList.Items) == 0 {
		return false, fmt.Errorf("no chain with name %s found", name)
	}

	return chainList.Items[0].Status.Phase == v1.PhaseRunning, nil
}
