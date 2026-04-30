// Package main wires up the Grafana Annotation Controller and starts the manager.
package main

import (
	"net/http"
	"os"
	goruntime "runtime"
	"strconv"
	"strings"
	"time"

	"github.com/perun-engineering/deployment-annotator-for-grafana/internal/controller"
	"github.com/perun-engineering/deployment-annotator-for-grafana/internal/grafana"
	"go.uber.org/zap/zapcore"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/metrics/server"
)

const httpTimeout = 30 * time.Second

var (
	version   = "dev"
	commit    = "unknown"
	buildTime = "unknown"
)

func main() {
	scheme := runtime.NewScheme()
	_ = corev1.AddToScheme(scheme)
	_ = appsv1.AddToScheme(scheme)

	ctrl.SetLogger(zap.New(zapOpts()...))
	logger := ctrl.Log.WithName("main")
	logger.Info("Grafana Annotation Controller",
		"version", version, "commit", commit, "buildTime", buildTime,
		"goVersion", goruntime.Version(), "os", goruntime.GOOS, "arch", goruntime.GOARCH)

	grafanaURL := requireEnv("GRAFANA_URL")
	grafanaKey := requireEnv("GRAFANA_API_KEY")

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                 scheme,
		Logger:                 ctrl.Log.WithName("manager"),
		Metrics:                server.Options{BindAddress: ":8081"},
		HealthProbeBindAddress: ":8080",
		LeaderElection:         false,
	})
	if err != nil {
		logger.Error(err, "Failed to create manager")
		os.Exit(1)
	}

	gc := &grafana.Client{
		URL:        strings.TrimSuffix(grafanaURL, "/"),
		APIKey:     grafanaKey,
		HTTPClient: &http.Client{Timeout: httpTimeout},
	}

	adapters := []struct {
		envKey  string
		adapter controller.WorkloadAdapter
	}{
		{"WATCH_DEPLOYMENTS", controller.DeploymentAdapter{}},
		{"WATCH_STATEFULSETS", controller.StatefulSetAdapter{}},
		{"WATCH_DAEMONSETS", controller.DaemonSetAdapter{}},
	}
	for _, a := range adapters {
		if !envBool(a.envKey, true) {
			logger.Info("Watching disabled", "kind", a.adapter.Kind())
			continue
		}
		r := &controller.WorkloadReconciler{
			Client:  mgr.GetClient(),
			Scheme:  mgr.GetScheme(),
			GClient: gc,
			Adapter: a.adapter,
		}
		if err := r.SetupWithManager(mgr); err != nil {
			logger.Error(err, "Failed to setup controller", "kind", a.adapter.Kind())
			os.Exit(1)
		}
	}

	_ = mgr.AddHealthzCheck("healthz", func(*http.Request) error { return nil })
	_ = mgr.AddReadyzCheck("readyz", func(*http.Request) error { return nil })

	logger.Info("Starting controller")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		logger.Error(err, "Manager exited with error")
		os.Exit(1)
	}
}

func requireEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		ctrl.Log.WithName("main").Error(nil, key+" is required")
		os.Exit(1)
	}
	return v
}

func envBool(key string, def bool) bool {
	if v := os.Getenv(key); v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			return b
		}
	}
	return def
}

func zapOpts() []zap.Opts {
	dev := envBool("LOG_DEVELOPMENT", false)
	opts := []zap.Opts{zap.UseDevMode(dev)}
	switch os.Getenv("LOG_LEVEL") {
	case "debug":
		opts = append(opts, zap.Level(zapcore.DebugLevel))
	case "error":
		opts = append(opts, zap.Level(zapcore.ErrorLevel))
	default:
		opts = append(opts, zap.Level(zapcore.InfoLevel))
	}
	return opts
}
