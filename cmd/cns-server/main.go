package main

import (
	k8srest "k8s.io/client-go/rest"

	"github.com/allinbits/demeris-backend/cns/chainwatch"
	"github.com/allinbits/demeris-backend/cns/database"
	"github.com/allinbits/demeris-backend/cns/rest"
	"github.com/allinbits/demeris-backend/utils/k8s"
	"github.com/allinbits/demeris-backend/utils/logging"
)

var Version = "not specified"

func main() {
	config, err := readConfig()
	if err != nil {
		panic(err)
	}

	logger := logging.New(logging.LoggingConfig{
		LogPath: config.LogPath,
		Debug:   config.Debug,
	})

	logger.Infow("cns-server", "version", Version)

	di, err := database.New(config.DatabaseConnectionURL)
	if err != nil {
		logger.Fatal(err)
	}

	kube, err := k8s.NewInCluster()
	if err != nil {
		logger.Fatal(err)
	}

	rc, err := chainwatch.NewConnection(config.Redis)
	if err != nil {
		logger.Fatal(err)
	}

	infConfig, err := k8srest.InClusterConfig()
	if err != nil {
		logger.Panicw("k8s server panic", "error", err)
	}

	nodesetInformer, err := k8s.GetInformer(infConfig, config.KubernetesNamespace, "nodesets")
	if err != nil {
		logger.Panicw("k8s server panic", "error", err)
	}

	relayerInformer, err := k8s.GetInformer(infConfig, config.KubernetesNamespace, "relayers")
	if err != nil {
		logger.Panicw("k8s server panic", "error", err)
	}

	go nodesetInformer.Informer().Run(make(chan struct{}))

	go relayerInformer.Informer().Run(make(chan struct{}))

	ci := chainwatch.New(
		logger,
		kube,
		config.KubernetesNamespace,
		rc,
		di,
		nodesetInformer,
		relayerInformer,
		config.RelayerDebug,
	)

	go ci.Run()

	restServer := rest.NewServer(
		logger,
		di,
		&kube,
		rc,
		config.KubernetesNamespace,
		nodesetInformer,
		config.Debug,
	)

	if err := restServer.Serve(config.RESTAddress); err != nil {
		logger.Panicw("rest http server error", "error", err)
	}
}
