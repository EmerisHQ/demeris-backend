package main

import (
	"net/http"
	"runtime"
	"runtime/debug"

	_ "net/http/pprof"

	"github.com/allinbits/demeris-backend/api/config"
	"github.com/allinbits/demeris-backend/api/database"
	"github.com/allinbits/demeris-backend/api/router"
	"github.com/allinbits/demeris-backend/utils"
	"github.com/allinbits/demeris-backend/utils/k8s"
	"github.com/allinbits/demeris-backend/utils/logging"
	"github.com/allinbits/demeris-backend/utils/store"
	gaia "github.com/cosmos/gaia/v5/app"
	_ "github.com/lib/pq"
	"k8s.io/client-go/rest"
)

var Version = "not specified"

func main() {
	cfg, err := config.Read()
	if err != nil {
		panic(err)
	}

	l := logging.New(logging.LoggingConfig{
		Debug: cfg.Debug,
	})

	if cfg.Debug {
		runtime.SetCPUProfileRate(500)

		go func() {
			http.HandleFunc("/freemem", func(_ http.ResponseWriter, _ *http.Request) {
				runtime.GC()
				debug.FreeOSMemory()
			})

			l.Debugw("starting profiling server", "port", "6060")
			err := http.ListenAndServe(":6060", nil)
			if err != nil {
				l.Panicw("cannot run profiling server", "error", err)
			}
		}()
	}

	l.Infow("api-server", "version", Version)

	dbi, err := database.Init(cfg)
	if err != nil {
		l.Panicw("cannot initialize database", "error", err)
	}

	s, err := store.NewClient(cfg.RedisAddr)
	if err != nil {
		l.Panicw("unable to start redis client", "error", err)
	}

	kubeClient, err := k8s.NewInCluster()
	if err != nil {
		l.Panicw("cannot initialize k8s", "error", err)
	}
	cdc, _ := gaia.MakeCodecs()

	r := router.New(
		dbi,
		l,
		s,
		kubeClient,
		cfg.KubernetesNamespace,
		cfg.CNSAddr,
		cdc,
		cfg.Debug,
	)

	infConfig, err := rest.InClusterConfig()
	if err != nil {
		l.Panicw("k8s server panic", "error", err)
	}

	informer, err := utils.GetInformer(infConfig, cfg.KubernetesNamespace, "relayers")
	if err != nil {
		l.Panicw("k8s server panic", "error", err)
	}

	go informer.Informer().Run(make(chan struct{}))

	if err := r.Serve(cfg.ListenAddr); err != nil {
		l.Panicw("http server panic", "error", err)
	}
}
