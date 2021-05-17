package operator

import (
	"fmt"

	v1 "github.com/allinbits/starport-operator/api/v1"
	corev1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

const (
	defaultMoniker   = "demeris"
	defaultNamespace = "default"

	tracelistenerImage = "demeris/tracelistener"

	trFifoPathVar = "TRACELISTENER_FIFOPATH"
	trFifoPath    = "/trace-store/kvstore.fifo"

	trDbURLVar = "TRACELISTENER_DATABASECONNECTIONURL"
	trDbURL    = "postgres://root@cockroachdb-public:26257?sslmode=disable"

	trTypeVar = "TRACELISTENER_TYPE"
	trType    = "gaia"

	trChainNameVar = "TRACELISTENER_CHAINNAME"
)

var (
	defaultStartupTimeout = "30m"
	defaultProtocol       = corev1.ProtocolTCP
	defaultPort           = intstr.FromInt(26257)
)

type NodeConfiguration struct {
	Name               string
	CLIName            string
	JoinConfig         *v1.JoinConfig
	TestnetConfig      *v1.ValidatorInitConfig
	DockerImage        string
	DockerImageVersion string
}

func (n NodeConfiguration) Validate() error {
	if n.JoinConfig == nil && n.TestnetConfig == nil {
		return fmt.Errorf("must specify either JoinConfing or TestnetConfig")
	}

	if n.Name == "" {
		return fmt.Errorf("missing name")
	}

	if n.DockerImage == "" {
		return fmt.Errorf("missing docker image")
	}

	if n.DockerImageVersion == "" {
		return fmt.Errorf("missing docker image version")
	}

	return nil
}

var DefaultNodeConfig = v1.NodeSet{
	TypeMeta: metav1.TypeMeta{
		Kind:       "NodeSet",
		APIVersion: "apps.starport.cloud/v1",
	},

	ObjectMeta: metav1.ObjectMeta{
		// Users must provide "Name" field
		Namespace: defaultNamespace,
	},
	Spec: v1.NodeSetSpec{
		Replicas: 1,
		App:      v1.AppDetails{
			// Users must provide "Name,DaemonName,CliName" field, they should probably all be
			// the same.
		},
		Moniker: defaultMoniker,
		Config: &v1.NodeSetConfig{
			Nodes: &v1.NodeSetConfigNodes{
				StartupTimeout: &defaultStartupTimeout,
				TraceStoreContainer: &v1.TraceStoreContainerConfig{
					Image:           tracelistenerImage,
					ImagePullPolicy: corev1.PullNever,
					Env: []corev1.EnvVar{
						{
							Name:  trFifoPathVar,
							Value: trFifoPath,
						},
						{
							Name:  trDbURLVar,
							Value: trDbURL,
						},
						{
							Name:  trTypeVar,
							Value: trType,
						},
					},
				},
			},
			AdditionalEgressRules: []netv1.NetworkPolicyEgressRule{
				{
					Ports: []netv1.NetworkPolicyPort{
						{
							Protocol: &defaultProtocol,
							Port:     &defaultPort,
						},
					},
				},
			},
		},
	},
	Status: v1.NodeSetStatus{},
}

func NewNode(c NodeConfiguration) (*v1.NodeSet, error) {
	if err := c.Validate(); err != nil {
		return nil, err
	}

	var node = DefaultNodeConfig

	node.ObjectMeta.Name = c.Name

	ns := &node.Spec
	ns.App.Name = c.Name
	ns.App.DaemonName = c.Name
	ns.App.CliName = c.Name
	if c.CLIName != "" {
		ns.App.CliName = c.CLIName
	}

	switch {
	case c.TestnetConfig != nil:
		ns.Init = c.TestnetConfig
	case c.JoinConfig != nil:
		ns.Join = c.JoinConfig
	}

	ns.Config.Nodes.TraceStoreContainer.Env = append(ns.Config.Nodes.TraceStoreContainer.Env, corev1.EnvVar{
		Name:  trChainNameVar,
		Value: c.Name,
	})

	return &node, nil
}
