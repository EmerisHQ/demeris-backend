package k8s

import (
	"fmt"
	"path/filepath"

	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/rest"

	v1 "github.com/allinbits/starport-operator/api/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// NewInCluster creates a new Kubernetes client to be used inside a cluster.
func NewInCluster() (client.Client, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, fmt.Errorf("cannot create in-cluster configuration, %w", err)
	}

	scheme := runtime.NewScheme()

	if err := v1.SchemeBuilder.AddToScheme(scheme); err != nil {
		return nil, fmt.Errorf("cannot add starport operator schemas, %w", err)
	}

	if err := corev1.SchemeBuilder.AddToScheme(scheme); err != nil {
		return nil, fmt.Errorf("cannot add core schemas, %w", err)
	}

	c, err := client.New(config, client.Options{
		Scheme: scheme,
	})

	if err != nil {
		return nil, fmt.Errorf("cannot create client, %w", err)
	}

	return c, nil
}

// New creates a new Kubernetes client to be used outside a cluster.
func New() (client.Client, error) {
	var kubeconfig string
	home := homedir.HomeDir()
	if home == "" {
		return nil, fmt.Errorf("kubernetes homedir empty")
	}

	kubeconfig = filepath.Join(home, ".kube", "config")

	// use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return nil, fmt.Errorf("cannot build configuration from flags, %w", err)
	}

	scheme := runtime.NewScheme()

	if err := v1.SchemeBuilder.AddToScheme(scheme); err != nil {
		return nil, fmt.Errorf("cannot add starport operator schemas, %w", err)
	}

	if err := corev1.SchemeBuilder.AddToScheme(scheme); err != nil {
		return nil, fmt.Errorf("cannot add core schemas, %w", err)
	}

	c, err := client.New(config, client.Options{
		Scheme: scheme,
	})

	if err != nil {
		return nil, fmt.Errorf("cannot create client, %w", err)
	}

	return c, nil
}
