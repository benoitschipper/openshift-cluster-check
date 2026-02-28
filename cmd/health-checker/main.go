// Command health-checker is an in-cluster OpenShift platform health checker.
// It periodically polls OpenShift/Kubernetes APIs and exposes five binary
// Prometheus gauges (0=healthy, 1=unhealthy) at /metrics.
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	openshiftclient "github.com/openshift/client-go/config/clientset/versioned"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"github.com/openshift-cluster-check/health-checker/internal/checker"
	"github.com/openshift-cluster-check/health-checker/internal/config"
	"github.com/openshift-cluster-check/health-checker/internal/metrics"
)

func main() {
	// 1. Parse configuration from environment variables.
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("ERROR: invalid configuration: %v", err)
	}

	log.Printf("Starting health-checker: interval=%s port=%d", cfg.CheckInterval, cfg.MetricsPort)

	// 2. Build in-cluster Kubernetes config.
	restCfg, err := rest.InClusterConfig()
	if err != nil {
		log.Fatalf("ERROR: failed to build in-cluster config: %v", err)
	}

	// 3. Create Kubernetes client.
	k8sClient, err := kubernetes.NewForConfig(restCfg)
	if err != nil {
		log.Fatalf("ERROR: failed to create Kubernetes client: %v", err)
	}

	// 4. Create OpenShift config client (for ClusterOperators and ClusterVersion).
	ocpClientset, err := openshiftclient.NewForConfig(restCfg)
	if err != nil {
		log.Fatalf("ERROR: failed to create OpenShift client: %v", err)
	}
	ocpClient := ocpClientset.ConfigV1()

	// 5. Register Prometheus metrics.
	metrics.Register()

	// 6. Set up context with OS signal handling for graceful shutdown.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		sig := <-sigCh
		log.Printf("Received signal %v, shutting down...", sig)
		cancel()
	}()

	// 7. Run one initial check cycle so metrics are populated before the first scrape.
	log.Println("Running initial health check cycle...")
	checker.RunChecks(ctx, k8sClient, ocpClient, cfg)

	// 8. Start the HTTP server for /metrics.
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "ok")
	})

	addr := fmt.Sprintf(":%d", cfg.MetricsPort)
	server := &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	go func() {
		log.Printf("Serving /metrics on %s", addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("ERROR: HTTP server failed: %v", err)
		}
	}()

	// 9. Start the periodic checker loop (blocks until context is cancelled).
	checker.StartLoop(ctx, k8sClient, ocpClient, cfg)

	// 10. Graceful HTTP server shutdown.
	shutdownCtx, shutdownCancel := context.WithCancel(context.Background())
	defer shutdownCancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("WARNING: HTTP server shutdown error: %v", err)
	}

	log.Println("health-checker stopped.")
}
