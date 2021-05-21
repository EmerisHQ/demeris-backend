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
	Name               string                  // we don't export this since the REST server will fill this for us
	CLIName            string                  `json:"cli_name"`
	JoinConfig         *v1.JoinConfig          `json:"join_config"`
	TestnetConfig      *v1.ValidatorInitConfig `json:"testnet_config"`
	DockerImage        string                  `json:"docker_image"`
	DockerImageVersion string                  `json:"docker_image_version"`
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
		Persistence: v1.NodesPersistenceSpec{
			Size: "5G",
		},
		SdkVersion: v1.Stargate,
		Moniker:    defaultMoniker,
	},
}

var defaultTracelistenerConfig = v1.TraceStoreContainerConfig{
	Image:           tracelistenerImage,
	ImagePullPolicy: corev1.PullIfNotPresent,
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
}

var defaultConfig = v1.NodeSetConfig{
	Nodes: &v1.NodeSetConfigNodes{
		StartupTimeout: &defaultStartupTimeout,
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
}

func NewNode(c NodeConfiguration) (*v1.NodeSet, error) {
	if err := c.Validate(); err != nil {
		return nil, err
	}

	var node = DefaultNodeConfig

	node.ObjectMeta.Name = c.Name

	ns := &node.Spec
	ns.App.Name = c.Name
	ns.App.CliName = c.Name
	ns.App.DaemonName = c.Name
	if c.CLIName != "" {
		ns.App.CliName = c.CLIName
		ns.App.DaemonName = c.CLIName
	}

	ns.Image = v1.Image{
		Name:    c.DockerImage,
		Version: c.DockerImageVersion,
	}

	switch {
	case c.TestnetConfig != nil:
		ns.Init = c.TestnetConfig
	case c.JoinConfig != nil:
		ns.Join = c.JoinConfig
	}

	tracelistenerConfig := defaultTracelistenerConfig

	tracelistenerConfig.Env = append(tracelistenerConfig.Env, corev1.EnvVar{
		Name:  trChainNameVar,
		Value: c.Name,
	})

	nodeConfig := defaultConfig
	nodeConfig.Nodes.TraceStoreContainer = &tracelistenerConfig

	node.Spec.Config = &nodeConfig

	return &node, nil
}
