package operator

import (
	"fmt"
	"strconv"
	"time"

	v1 "github.com/allinbits/starport-operator/api/v1"
	corev1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

const (
	defaultMoniker   = "emeris"
	defaultNamespace = "emeris"

	tracelistenerImage = "emeris/tracelistener"

	trFifoPathVar = "TRACELISTENER_FIFOPATH"
	trFifoPath    = "/trace-store/kvstore.fifo"

	trDbURLVar = "TRACELISTENER_DATABASECONNECTIONURL"
	trDbURL    = "postgres://root@cockroachdb-public:26257?sslmode=disable"

	trTypeVar = "TRACELISTENER_TYPE"
	trType    = "gaia"

	trChainNameVar = "TRACELISTENER_CHAINNAME"
	trDebugVar     = "TRACELISTENER_DEBUG"
)

var (
	defaultStartupTimeout = "30m"
	defaultProtocol       = corev1.ProtocolTCP
	defaultPort           = intstr.FromInt(26257)

	DefaultRelayerConfiguration = RelayerConfiguration{
		MaxMsgNum:      15,
		MaxGas:         500000,
		ClockDrift:     "1800s",
		TrustingPeriod: "14days",
	}
)

type NodeConfiguration struct {
	Name                string                  // we don't export this since the REST server will fill this for us
	CLIName             string                  `json:"cli_name"`
	JoinConfig          *v1.JoinConfig          `json:"join_config"`
	TestnetConfig       *v1.ValidatorInitConfig `json:"testnet_config"`
	DockerImage         string                  `json:"docker_image"`
	DockerImageVersion  string                  `json:"docker_image_version"`
	Namespace           string                  `json:"-"`
	TracelistenerImage  string                  `json:"tracelistener_image"`
	DisableMinFeeConfig bool                    `json:"disable_min_fee_config"`
	TracelistenerDebug  bool
}

type RelayerConfiguration struct {
	MaxMsgNum      int64  `json:"max_msg_num"`
	MaxGas         int64  `json:"max_gas"`
	ClockDrift     string `json:"clock_drift"`
	TrustingPeriod string `json:"trusting_period"`
}

func (r RelayerConfiguration) Validate() error {
	if r.MaxMsgNum <= 0 {
		return fmt.Errorf("max msg num can't be less than or equal to zero")
	}

	if r.MaxGas <= 0 {
		return fmt.Errorf("max gas can't be less than or equal to zero")
	}

	if _, err := time.ParseDuration(r.ClockDrift); err != nil {
		return fmt.Errorf("cannot parse clock drift expression, %w", err)
	}

	// we can't parse trusting period :/
	return nil
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

	if n.Namespace == "" {
		return fmt.Errorf("missing namespace")
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
		Moniker: defaultMoniker,
	},
}

var defaultTracelistenerConfig = v1.TraceStoreContainerConfig{
	Image:           tracelistenerImage,
	ImagePullPolicy: corev1.PullAlways,
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

	if c.Namespace == "" {
		c.Namespace = defaultNamespace
	}

	if c.TracelistenerImage == "" {
		c.TracelistenerImage = tracelistenerImage
	}

	var node = DefaultNodeConfig

	node.ObjectMeta.Namespace = c.Namespace

	node.ObjectMeta.Name = c.Name

	ns := &node.Spec
	ns.App.Name = c.Name
	ns.App.CliName = &c.Name
	ns.App.DaemonName = &c.Name
	if c.CLIName != "" {
		ns.App.CliName = &c.CLIName
		ns.App.DaemonName = &c.CLIName
	}

	ns.Image = v1.Image{
		Name:    c.DockerImage,
		Version: &c.DockerImageVersion,
	}

	switch {
	case c.TestnetConfig != nil:
		ns.Init = c.TestnetConfig
	case c.JoinConfig != nil:
		ns.Join = c.JoinConfig
	}

	tracelistenerConfig := defaultTracelistenerConfig

	tracelistenerConfig.Image = c.TracelistenerImage

	tracelistenerConfig.Env = append(tracelistenerConfig.Env, corev1.EnvVar{
		Name:  trChainNameVar,
		Value: c.Name,
	})

	tracelistenerConfig.Env = append(tracelistenerConfig.Env, corev1.EnvVar{
		Name:  trDebugVar,
		Value: strconv.FormatBool(c.TracelistenerDebug),
	})

	nodeConfig := defaultConfig
	nodeConfig.Nodes.TraceStoreContainer = &tracelistenerConfig

	node.Spec.Config = &nodeConfig

	return &node, nil
}
