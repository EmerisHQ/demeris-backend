package k8s

import (
	"context"

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
