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
