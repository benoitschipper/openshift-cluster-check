package checker

import (
	"context"
	"log"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/openshift-cluster-check/health-checker/internal/config"
	"github.com/openshift-cluster-check/health-checker/internal/metrics"
)

// fatalContainerReasons are container waiting/terminated reasons that indicate a fatal failure.
var fatalContainerReasons = map[string]bool{
	"CrashLoopBackOff": true,
	"OOMKilled":        true,
	"Error":            true,
}

// CheckSystemPods lists pods in each system namespace and updates openshift_system_pods_failing:
//   - 1 if any pod has phase=Failed or any container has a fatal reason (CrashLoopBackOff, OOMKilled, Error)
//   - 0 if all system pods are healthy
//
// Pods are listed per namespace (not cluster-wide) to minimize API server load.
// On API error for any namespace, the metric is set to 1 (fail-closed).
func CheckSystemPods(ctx context.Context, client kubernetes.Interface, cfg config.Config) {
	// Collect all system namespaces by listing all namespaces and filtering
	nsList, err := client.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		log.Printf("WARNING: failed to list Namespaces: %v — setting metric to 1 (fail-closed)", err)
		metrics.SystemPodsFailing.Set(1)
		return
	}

	failing := false
	for _, ns := range nsList.Items {
		if !IsSystemNamespace(ns.Name, cfg) {
			continue
		}

		pods, err := client.CoreV1().Pods(ns.Name).List(ctx, metav1.ListOptions{})
		if err != nil {
			log.Printf("WARNING: failed to list Pods in namespace %q: %v — setting metric to 1 (fail-closed)", ns.Name, err)
			metrics.SystemPodsFailing.Set(1)
			return
		}

		for _, pod := range pods.Items {
			if isPodFailing(pod) {
				failing = true
				log.Printf("WARNING: Pod %q in namespace %q is failing", pod.Name, pod.Namespace)
			}
		}
	}

	if failing {
		metrics.SystemPodsFailing.Set(1)
	} else {
		metrics.SystemPodsFailing.Set(0)
	}
}

// isPodFailing returns true if the pod has phase=Failed or any container
// has a fatal waiting or terminated reason.
func isPodFailing(pod corev1.Pod) bool {
	// Check pod phase
	if pod.Status.Phase == corev1.PodFailed {
		return true
	}

	// Check container statuses for fatal states
	for _, cs := range pod.Status.ContainerStatuses {
		if isContainerFailing(cs) {
			return true
		}
	}

	// Check init container statuses
	for _, cs := range pod.Status.InitContainerStatuses {
		if isContainerFailing(cs) {
			return true
		}
	}

	return false
}

// isContainerFailing returns true if the container has a fatal waiting or terminated reason.
func isContainerFailing(cs corev1.ContainerStatus) bool {
	if cs.State.Waiting != nil && fatalContainerReasons[cs.State.Waiting.Reason] {
		return true
	}
	if cs.State.Terminated != nil && fatalContainerReasons[cs.State.Terminated.Reason] {
		return true
	}
	return false
}
