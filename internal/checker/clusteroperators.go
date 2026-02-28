package checker

import (
	"context"
	"fmt"
	"log"

	configv1 "github.com/openshift/api/config/v1"
	configv1client "github.com/openshift/client-go/config/clientset/versioned/typed/config/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/openshift-cluster-check/health-checker/internal/metrics"
)

// CheckClusterOperators lists all ClusterOperators and updates:
//   - openshift_cluster_operators_degraded: 1 if any operator (excluding etcd) is degraded or unavailable
//   - openshift_etcd_degraded: 1 if the etcd operator is degraded or unavailable
//
// On API error, both metrics are set to 1 (fail-closed).
func CheckClusterOperators(ctx context.Context, client configv1client.ConfigV1Interface) {
	operators, err := client.ClusterOperators().List(ctx, metav1.ListOptions{})
	if err != nil {
		log.Printf("WARNING: failed to list ClusterOperators: %v — setting both metrics to 1 (fail-closed)", err)
		metrics.ClusterOperatorsDegraded.Set(1)
		metrics.EtcdDegraded.Set(1)
		return
	}

	operatorsDegraded := false
	etcdDegraded := false

	for _, op := range operators.Items {
		degraded := isOperatorDegraded(op)
		if op.Name == "etcd" {
			if degraded {
				etcdDegraded = true
				log.Printf("WARNING: ClusterOperator %q is degraded or unavailable", op.Name)
			}
		} else {
			if degraded {
				operatorsDegraded = true
				log.Printf("WARNING: ClusterOperator %q is degraded or unavailable", op.Name)
			}
		}
	}

	if operatorsDegraded {
		metrics.ClusterOperatorsDegraded.Set(1)
	} else {
		metrics.ClusterOperatorsDegraded.Set(0)
	}

	if etcdDegraded {
		metrics.EtcdDegraded.Set(1)
	} else {
		metrics.EtcdDegraded.Set(0)
	}
}

// isOperatorDegraded returns true if the ClusterOperator has Degraded=True or Available=False.
func isOperatorDegraded(op configv1.ClusterOperator) bool {
	for _, cond := range op.Status.Conditions {
		switch cond.Type {
		case configv1.OperatorDegraded:
			if cond.Status == configv1.ConditionTrue {
				return true
			}
		case configv1.OperatorAvailable:
			if cond.Status == configv1.ConditionFalse {
				return true
			}
		}
	}
	return false
}

// clusterOperatorConditionStatus returns the status of a named condition, or empty string if not found.
func clusterOperatorConditionStatus(op configv1.ClusterOperator, condType configv1.ClusterStatusConditionType) configv1.ConditionStatus {
	for _, cond := range op.Status.Conditions {
		if cond.Type == condType {
			return cond.Status
		}
	}
	return ""
}

// clusterOperatorConditionMessage returns the message of a named condition, or empty string if not found.
func clusterOperatorConditionMessage(op configv1.ClusterOperator, condType configv1.ClusterStatusConditionType) string {
	for _, cond := range op.Status.Conditions {
		if cond.Type == condType {
			return cond.Message
		}
	}
	return ""
}

// Ensure these helpers are used (suppress unused warnings if needed).
var _ = fmt.Sprintf
var _ = clusterOperatorConditionStatus
var _ = clusterOperatorConditionMessage
