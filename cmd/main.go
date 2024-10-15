package main

import (
	"log/slog"
	"net/http"
	"os"

	_ "net/http/pprof"

	"github.com/go-chi/chi/v5"
	"github.com/k8scope/k8s-restart-app/internal/api"
	"github.com/k8scope/k8s-restart-app/internal/config"
	"github.com/k8scope/k8s-restart-app/internal/ledger"
	"github.com/k8scope/k8s-restart-app/internal/lock"
	"github.com/k8scope/k8s-restart-app/internal/utils"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	envListenAddress           = utils.StringEnvOrDefault("LISTEN_ADDRESS", ":8080")
	envListenMonitoringAddress = utils.StringEnvOrDefault("LISTEN_MONITORING_ADDRESS", ":6060")
	envConfigFilePath          = utils.StringEnvOrDefault("CONFIG_FILE_PATH", "config.yaml")
	envKubeConfigPath          = utils.StringEnvOrDefault("KUBE_CONFIG_PATH", "")
	envWatchInterval           = utils.IntEnvOrDefault("WATCH_INTERVAL", 10)
	envForceUnlockSec          = utils.IntEnvOrDefault("FORCE_UNLOCK_SEC", 300)

	// non env variables
	k8sClient *kubernetes.Clientset
	// lock handling
	lockH *lock.Lock = lock.NewLock(lock.NewInMem(), envForceUnlockSec)

	appConfig *config.Config

	ldgr *ledger.Ledger
)

func init() {
	// setup K8s client
	var k8sConfig *rest.Config
	if envKubeConfigPath != "" {
		cfg, err := clientcmd.BuildConfigFromFlags("", envKubeConfigPath)
		if err != nil {
			slog.Error("failed to get kubeconfig", "error", err)
			os.Exit(-1)
		}
		k8sConfig = cfg
	} else {
		cfg, err := rest.InClusterConfig()
		if err != nil {
			slog.Error("failed to get in-cluster config", "error", err)
			os.Exit(-1)
		}
		k8sConfig = cfg
	}
	clientset, err := kubernetes.NewForConfig(k8sConfig)
	if err != nil {
		slog.Error("failed to create k8s client", "error", err)
		os.Exit(-1)
	}
	k8sClient = clientset

	// load config
	cfg, err := config.ReadConfigFile(envConfigFilePath)
	if err != nil {
		slog.Error("failed to read config file", "error", err)
		os.Exit(-1)
	}
	appConfig = cfg

	// setup ledger and watch apps
	ldgr = ledger.New(k8sClient, lockH, envWatchInterval)
	for _, app := range appConfig.Services {
		ldgr.Watch(app)
	}

}

func main() {
	defer ldgr.Close()
	slog.Info("starting server...", "listen_address", envListenAddress)

	rt := chi.NewRouter()
	rt.Get("/", api.Index)
	rt.Route("/api/v1", func(r chi.Router) {
		r.Route("/service", func(r chi.Router) {
			r.Get("/", api.ListApplications(*appConfig))
			r.Get("/status", api.Status(ldgr))
			r.Route("/{kind}/{namespace}/{name}", func(r chi.Router) {
				r.Use(api.MiddlewareValidation(*appConfig))
				r.Post("/restart", api.Restart(k8sClient, lockH))
			})
		})
	})

	go func() {
		http.Handle("/metrics", promhttp.Handler())
		slog.Info("starting monitoring server...", "listen_address", envListenMonitoringAddress)
		err := http.ListenAndServe(envListenMonitoringAddress, nil)
		if err != nil {
			slog.Error("failed to start monitoring server", "error", err)
			os.Exit(-1)
		}
	}()

	err := http.ListenAndServe(envListenAddress, rt)
	if err != nil {
		slog.Error("failed to start server", "error", err)
		os.Exit(-1)
	}
}
