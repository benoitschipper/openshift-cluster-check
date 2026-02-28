package checker

import (
	"testing"

	"github.com/openshift-cluster-check/health-checker/internal/config"
)

func defaultConfig() config.Config {
	// Matches the new defaults: openshift- and kube- prefixes; no exact-match list.
	return config.Config{
		SystemNamespacePrefixes: []string{"openshift-", "kube-"},
		SystemNamespaces:        []string{},
	}
}

func TestIsSystemNamespace_OpenshiftPrefix(t *testing.T) {
	cfg := defaultConfig()
	if !IsSystemNamespace("openshift-monitoring", cfg) {
		t.Error("expected openshift-monitoring to be a system namespace")
	}
}

// kube-* namespaces are now matched via the kube- prefix rule, not exact-match.

func TestIsSystemNamespace_KubeSystem_ViaPrefix(t *testing.T) {
	cfg := defaultConfig()
	if !IsSystemNamespace("kube-system", cfg) {
		t.Error("expected kube-system to be a system namespace via kube- prefix")
	}
}

func TestIsSystemNamespace_KubePublic_ViaPrefix(t *testing.T) {
	cfg := defaultConfig()
	if !IsSystemNamespace("kube-public", cfg) {
		t.Error("expected kube-public to be a system namespace via kube- prefix")
	}
}

func TestIsSystemNamespace_KubeNodeLease_ViaPrefix(t *testing.T) {
	cfg := defaultConfig()
	if !IsSystemNamespace("kube-node-lease", cfg) {
		t.Error("expected kube-node-lease to be a system namespace via kube- prefix")
	}
}

func TestIsSystemNamespace_KubeFutureSystem_ViaPrefix(t *testing.T) {
	cfg := defaultConfig()
	if !IsSystemNamespace("kube-future-system", cfg) {
		t.Error("expected kube-future-system to be a system namespace via kube- prefix")
	}
}

func TestIsSystemNamespace_KubernetesOps_NotSystem(t *testing.T) {
	cfg := defaultConfig()
	// "kubernetes-ops" starts with "kubernetes-", not "kube-", so it must NOT match.
	if IsSystemNamespace("kubernetes-ops", cfg) {
		t.Error("expected kubernetes-ops to NOT be a system namespace (does not start with kube-)")
	}
}

func TestIsSystemNamespace_Default_NotSystem(t *testing.T) {
	cfg := defaultConfig()
	if IsSystemNamespace("default", cfg) {
		t.Error("expected default to NOT be a system namespace")
	}
}

func TestIsSystemNamespace_MyApp_NotSystem(t *testing.T) {
	cfg := defaultConfig()
	if IsSystemNamespace("my-app", cfg) {
		t.Error("expected my-app to NOT be a system namespace")
	}
}

func TestIsSystemNamespace_ProductionApp_NotSystem(t *testing.T) {
	cfg := defaultConfig()
	if IsSystemNamespace("production-app", cfg) {
		t.Error("expected production-app to NOT be a system namespace")
	}
}

func TestIsSystemNamespace_CustomPrefix(t *testing.T) {
	cfg := config.Config{
		SystemNamespacePrefixes: []string{"openshift-", "my-infra-"},
		SystemNamespaces:        []string{"kube-system", "kube-public", "kube-node-lease"},
	}
	if !IsSystemNamespace("my-infra-logging", cfg) {
		t.Error("expected my-infra-logging to be a system namespace with custom prefix")
	}
}

func TestIsSystemNamespace_CustomExactMatch(t *testing.T) {
	cfg := config.Config{
		SystemNamespacePrefixes: []string{"openshift-"},
		SystemNamespaces:        []string{"kube-system", "kube-public", "kube-node-lease", "monitoring"},
	}
	if !IsSystemNamespace("monitoring", cfg) {
		t.Error("expected monitoring to be a system namespace with custom exact match")
	}
}

func TestIsSystemNamespace_OpenshiftEtcd(t *testing.T) {
	cfg := defaultConfig()
	if !IsSystemNamespace("openshift-etcd", cfg) {
		t.Error("expected openshift-etcd to be a system namespace")
	}
}
