package checker

import (
	"context"
	"log"

	configv1 "github.com/openshift/api/config/v1"
	configv1client "github.com/openshift/client-go/config/clientset/versioned/typed/config/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/openshift-cluster-check/health-checker/internal/metrics"
)

// CheckClusterVersion gets the ClusterVersion named "version" and updates
// openshift_clusterversion_degraded: 1 if Degraded=True or Available=False, 0 otherwise.
//
// On API error, the metric is set to 1 (fail-closed).
func CheckClusterVersion(ctx context.Context, client configv1client.ConfigV1Interface) {
	cv, err := client.ClusterVersions().Get(ctx, "version", metav1.GetOptions{})
	if err != nil {
		log.Printf("WARNING: failed to get ClusterVersion 'version': %v — setting metric to 1 (fail-closed)", err)
		metrics.ClusterVersionDegraded.Set(1)
		return
	}

	if isClusterVersionDegraded(*cv) {
		metrics.ClusterVersionDegraded.Set(1)
	} else {
		metrics.ClusterVersionDegraded.Set(0)
	}
}

// isClusterVersionDegraded returns true if the ClusterVersion has Degraded=True or Available=False.
func isClusterVersionDegraded(cv configv1.ClusterVersion) bool {
	for _, cond := range cv.Status.Conditions {
		switch cond.Type {
		case configv1.OperatorDegraded:
			if cond.Status == configv1.ConditionTrue {
				log.Printf("WARNING: ClusterVersion 'version' is Degraded: %s", cond.Message)
				return true
			}
		case configv1.OperatorAvailable:
			if cond.Status == configv1.ConditionFalse {
				log.Printf("WARNING: ClusterVersion 'version' is not Available: %s", cond.Message)
				return true
			}
		}
	}
	return false
}
