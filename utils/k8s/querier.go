package k8s

import (
	"context"
	"fmt"

	v1 "github.com/allinbits/starport-operator/api/v1"

	kube "sigs.k8s.io/controller-runtime/pkg/client"
)

var ErrNotFound = fmt.Errorf("not found")

type Querier struct {
	Namespace string
	Client    kube.Client
}

func (q Querier) ChainRunning(name string) (bool, error) {
	var chainList v1.NodeSetList

	if err := q.Client.List(context.TODO(), &chainList, kube.MatchingFields{
		"metadata.name": name,
	}, kube.InNamespace(q.Namespace)); err != nil {
		return false, err
	}

	if len(chainList.Items) == 0 {
		return false, fmt.Errorf("no chain with name %s found", name)
	}

	return chainList.Items[0].Status.Phase == v1.PhaseRunning, nil
}

func (q Querier) ChainByName(name string) (v1.NodeSet, error) {
	var chainList v1.NodeSetList

	if err := q.Client.List(context.TODO(), &chainList, kube.MatchingFields{
		"metadata.name": name,
	}, kube.InNamespace(q.Namespace)); err != nil {
		return v1.NodeSet{}, err
	}

	if len(chainList.Items) == 0 {
		return v1.NodeSet{}, fmt.Errorf("%w: %s", ErrNotFound, name)
	}

	return chainList.Items[0], nil
}

func (q Querier) ChainsByName(names ...string) ([]v1.NodeSet, error) {
	var chainList []v1.NodeSet

	for _, name := range names {
		c, err := q.ChainByName(name)
		if err != nil {
			return nil, err
		}

		chainList = append(chainList, c)
	}

	return chainList, nil
}

func (q Querier) AddNode(node v1.NodeSet) error {
	if err := q.Client.Create(context.TODO(), &node); err != nil {
		return err
	}

	return nil
}

func (q Querier) DeleteNode(nodeName string) error {
	objs := v1.NodeSetList{}
	if err := q.Client.List(context.TODO(), &objs, kube.MatchingFields{
		"metadata.name": nodeName,
	}, kube.InNamespace(q.Namespace)); err != nil {
		return err
	}

	if len(objs.Items) == 0 {
		return fmt.Errorf("%w: %s", ErrNotFound, nodeName)
	}

	if err := q.Client.Delete(context.TODO(), &objs.Items[0]); err != nil {
		return err
	}

	return nil
}

func (q Querier) Relayer() (v1.Relayer, error) {
	var e v1.RelayerList

	if err := q.Client.List(context.TODO(), &e, kube.InNamespace(q.Namespace)); err != nil {
		return v1.Relayer{}, err
	}

	if len(e.Items) == 0 {
		return v1.Relayer{}, ErrNotFound
	}

	return e.Items[0], nil
}

func (q Querier) AddRelayer(r v1.Relayer) error {
	if err := q.Client.Create(context.TODO(), &r); err != nil {
		return err
	}

	return nil
}

func (q Querier) UpdateRelayer(r v1.Relayer) error {
	if err := q.Client.Update(context.TODO(), &r); err != nil {
		return err
	}

	return nil
}
