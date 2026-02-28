// Package metrics defines and registers the five Prometheus binary gauge metrics
// exposed by the OpenShift cluster health-checker.
package metrics

import "github.com/prometheus/client_golang/prometheus"

// All five binary gauges: 0 = healthy, 1 = unhealthy.
var (
	// ClusterOperatorsDegraded is set to 1 if any ClusterOperator (excluding etcd)
	// has Degraded=True or Available=False.
	ClusterOperatorsDegraded = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "openshift_cluster_operators_degraded",
		Help: "1 if any ClusterOperator (excluding etcd) is degraded or unavailable, 0 otherwise.",
	})

	// NodesNotReady is set to 1 if any Node has condition Ready != True.
	NodesNotReady = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "openshift_nodes_not_ready",
		Help: "1 if any Node has condition Ready != True, 0 otherwise.",
	})

	// SystemPodsFailing is set to 1 if any pod in a system namespace has
	// phase=Failed or a container in CrashLoopBackOff, OOMKilled, or Error state.
	SystemPodsFailing = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "openshift_system_pods_failing",
		Help: "1 if any pod in a system/platform namespace is failing (phase=Failed or container in CrashLoopBackOff/OOMKilled/Error), 0 otherwise.",
	})

	// ClusterVersionDegraded is set to 1 if the ClusterVersion named "version"
	// has Degraded=True or Available=False.
	ClusterVersionDegraded = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "openshift_clusterversion_degraded",
		Help: "1 if the ClusterVersion 'version' is degraded or unavailable, 0 otherwise.",
	})

	// EtcdDegraded is set to 1 if the ClusterOperator named "etcd"
	// has Degraded=True or Available=False.
	EtcdDegraded = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "openshift_etcd_degraded",
		Help: "1 if the etcd ClusterOperator is degraded or unavailable, 0 otherwise.",
	})
)

// Register registers all five gauges with the default Prometheus registry.
// This should be called once at startup before the /metrics endpoint is served.
func Register() {
	prometheus.MustRegister(
		ClusterOperatorsDegraded,
		NodesNotReady,
		SystemPodsFailing,
		ClusterVersionDegraded,
		EtcdDegraded,
	)
}
