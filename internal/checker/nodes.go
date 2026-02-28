package checker

import (
	"context"
	"log"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/openshift-cluster-check/health-checker/internal/metrics"
)

// CheckNodes lists all Nodes and updates openshift_nodes_not_ready:
//   - 1 if any node has condition Ready != True
//   - 0 if all nodes are ready
//
// On API error, the metric is set to 1 (fail-closed).
func CheckNodes(ctx context.Context, client kubernetes.Interface) {
	nodes, err := client.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		log.Printf("WARNING: failed to list Nodes: %v — setting metric to 1 (fail-closed)", err)
		metrics.NodesNotReady.Set(1)
		return
	}

	notReady := false
	for _, node := range nodes.Items {
		if !isNodeReady(node) {
			notReady = true
			log.Printf("WARNING: Node %q is not Ready", node.Name)
		}
	}

	if notReady {
		metrics.NodesNotReady.Set(1)
	} else {
		metrics.NodesNotReady.Set(0)
	}
}

// isNodeReady returns true if the node has condition Ready=True.
// A node without a Ready condition is considered not ready.
func isNodeReady(node corev1.Node) bool {
	for _, cond := range node.Status.Conditions {
		if cond.Type == corev1.NodeReady {
			return cond.Status == corev1.ConditionTrue
		}
	}
	// No Ready condition found — treat as not ready
	return false
}
