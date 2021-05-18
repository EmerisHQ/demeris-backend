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
	}); err != nil {
		return err
	}

	if len(objs.Items) == 0 {
		return fmt.Errorf("node with name %s not found", nodeName)
	}

	if err := q.Client.Delete(context.TODO(), &objs.Items[0]); err != nil {
		return err
	}

	return nil
}
