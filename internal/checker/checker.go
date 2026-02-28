package checker

import (
	"context"
	"log"
	"time"

	configv1client "github.com/openshift/client-go/config/clientset/versioned/typed/config/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/openshift-cluster-check/health-checker/internal/config"
)

// RunChecks executes all five health checks independently.
// A failure in one check does not prevent the others from running.
// Each check updates its own Prometheus metric.
func RunChecks(ctx context.Context, k8sClient kubernetes.Interface, ocpClient configv1client.ConfigV1Interface, cfg config.Config) {
	log.Println("Running health checks...")

	// Each check is called independently; errors are handled internally per check.
	CheckClusterOperators(ctx, ocpClient)
	CheckClusterVersion(ctx, ocpClient)
	CheckNodes(ctx, k8sClient)
	CheckSystemPods(ctx, k8sClient, cfg)

	log.Println("Health checks complete.")
}

// StartLoop runs RunChecks on the configured interval using a ticker.
// It blocks until the context is cancelled.
// An initial check is NOT run here — callers should call RunChecks once before
// starting the HTTP server, then call StartLoop for subsequent periodic checks.
func StartLoop(ctx context.Context, k8sClient kubernetes.Interface, ocpClient configv1client.ConfigV1Interface, cfg config.Config) {
	ticker := time.NewTicker(cfg.CheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("Checker loop stopping: context cancelled.")
			return
		case <-ticker.C:
			RunChecks(ctx, k8sClient, ocpClient, cfg)
		}
	}
}
