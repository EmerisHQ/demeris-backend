package k8s

import (
	"fmt"

	appsv1 "github.com/allinbits/starport-operator/api/v1"
	"github.com/cosmos/cosmos-sdk/types/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	k8stypes "k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/dynamicinformer"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/rest"
)

func GetInformer(cfg *rest.Config, namespace, resourceType string) (informers.GenericInformer, error) {
	dc, err := dynamic.NewForConfig(cfg)
	if err != nil {
		return nil, err
	}

	factory := dynamicinformer.NewFilteredDynamicSharedInformerFactory(dc, 0, namespace, nil)
	return factory.ForResource(schema.GroupVersionResource{
		Group:    appsv1.GroupVersion.Group,
		Version:  appsv1.GroupVersion.Version,
		Resource: resourceType,
	}), nil
}

func GetChain(informer informers.GenericInformer, namespace, name string) (appsv1.NodeSet, error) {
	var chainList appsv1.NodeSetList
	obj, err := informer.Lister().Get(k8stypes.NamespacedName{
		Namespace: namespace,
		Name:      name,
	}.String())

	if err != nil {
		return appsv1.NodeSet{}, err
	}

	err = runtime.DefaultUnstructuredConverter.FromUnstructured(
		obj.(*unstructured.Unstructured).UnstructuredContent(), &chainList)
	if err != nil {
		return appsv1.NodeSet{}, errors.Wrap(err, "this is from unstructure")
	}

	if len(chainList.Items) == 0 {
		return appsv1.NodeSet{}, fmt.Errorf("%w: %s", ErrNotFound, name)
	}
	return chainList.Items[0], nil
}

func GetRelayer(informer informers.GenericInformer, namespace string) (appsv1.Relayer, error) {
	var relayer appsv1.Relayer
	obj, err := informer.Lister().Get(k8stypes.NamespacedName{
		Namespace: namespace,
		Name:      "relayer",
	}.String())

	if err != nil {
		return appsv1.Relayer{}, err
	}

	err = runtime.DefaultUnstructuredConverter.FromUnstructured(
		obj.(*unstructured.Unstructured).UnstructuredContent(), relayer)
	if err != nil {
		return appsv1.Relayer{}, err
	}

	return relayer, nil
}

func ChainRunning(informer informers.GenericInformer, namespace, name string) (bool, error) {
	var chainList appsv1.NodeSetList
	obj, err := informer.Lister().Get(k8stypes.NamespacedName{
		Namespace: namespace,
		Name:      name,
	}.String())

	if err != nil {
		return false, err
	}

	err = runtime.DefaultUnstructuredConverter.FromUnstructured(
		obj.(*unstructured.Unstructured).UnstructuredContent(), chainList)
	if err != nil {
		return false, err
	}

	if len(chainList.Items) == 0 {
		return false, fmt.Errorf("no chain with name %s found", name)
	}

	return chainList.Items[0].Status.Phase == appsv1.PhaseRunning, nil
}
