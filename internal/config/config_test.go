package config

import (
	"testing"
	"time"
)

func TestLoad_Defaults(t *testing.T) {
	// Clear all relevant env vars
	t.Setenv("CHECK_INTERVAL", "")
	t.Setenv("METRICS_PORT", "")
	t.Setenv("SYSTEM_NAMESPACE_PREFIXES", "")
	t.Setenv("SYSTEM_NAMESPACES", "")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if cfg.CheckInterval != 30*time.Second {
		t.Errorf("expected CheckInterval=30s, got %v", cfg.CheckInterval)
	}
	if cfg.MetricsPort != 8080 {
		t.Errorf("expected MetricsPort=8080, got %d", cfg.MetricsPort)
	}
	expectedPrefixes := []string{"openshift-", "kube-"}
	if len(cfg.SystemNamespacePrefixes) != len(expectedPrefixes) {
		t.Errorf("expected SystemNamespacePrefixes=%v, got %v", expectedPrefixes, cfg.SystemNamespacePrefixes)
	}
	for i, p := range expectedPrefixes {
		if cfg.SystemNamespacePrefixes[i] != p {
			t.Errorf("expected SystemNamespacePrefixes[%d]=%q, got %q", i, p, cfg.SystemNamespacePrefixes[i])
		}
	}
	// Default SYSTEM_NAMESPACES is empty — the kube- prefix subsumes all kube-* namespaces.
	if len(cfg.SystemNamespaces) != 0 {
		t.Errorf("expected SystemNamespaces=[], got %v", cfg.SystemNamespaces)
	}
}

func TestLoad_CustomCheckInterval(t *testing.T) {
	t.Setenv("CHECK_INTERVAL", "60")
	t.Setenv("METRICS_PORT", "")
	t.Setenv("SYSTEM_NAMESPACE_PREFIXES", "")
	t.Setenv("SYSTEM_NAMESPACES", "")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if cfg.CheckInterval != 60*time.Second {
		t.Errorf("expected CheckInterval=60s, got %v", cfg.CheckInterval)
	}
}

func TestLoad_InvalidCheckInterval(t *testing.T) {
	t.Setenv("CHECK_INTERVAL", "abc")
	t.Setenv("METRICS_PORT", "")
	t.Setenv("SYSTEM_NAMESPACE_PREFIXES", "")
	t.Setenv("SYSTEM_NAMESPACES", "")

	_, err := Load()
	if err == nil {
		t.Fatal("expected error for invalid CHECK_INTERVAL, got nil")
	}
}

func TestLoad_ZeroCheckInterval(t *testing.T) {
	t.Setenv("CHECK_INTERVAL", "0")
	t.Setenv("METRICS_PORT", "")
	t.Setenv("SYSTEM_NAMESPACE_PREFIXES", "")
	t.Setenv("SYSTEM_NAMESPACES", "")

	_, err := Load()
	if err == nil {
		t.Fatal("expected error for CHECK_INTERVAL=0, got nil")
	}
}

func TestLoad_NegativeCheckInterval(t *testing.T) {
	t.Setenv("CHECK_INTERVAL", "-5")
	t.Setenv("METRICS_PORT", "")
	t.Setenv("SYSTEM_NAMESPACE_PREFIXES", "")
	t.Setenv("SYSTEM_NAMESPACES", "")

	_, err := Load()
	if err == nil {
		t.Fatal("expected error for negative CHECK_INTERVAL, got nil")
	}
}

func TestLoad_CustomMetricsPort(t *testing.T) {
	t.Setenv("CHECK_INTERVAL", "")
	t.Setenv("METRICS_PORT", "9090")
	t.Setenv("SYSTEM_NAMESPACE_PREFIXES", "")
	t.Setenv("SYSTEM_NAMESPACES", "")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if cfg.MetricsPort != 9090 {
		t.Errorf("expected MetricsPort=9090, got %d", cfg.MetricsPort)
	}
}

func TestLoad_InvalidMetricsPort(t *testing.T) {
	t.Setenv("CHECK_INTERVAL", "")
	t.Setenv("METRICS_PORT", "notaport")
	t.Setenv("SYSTEM_NAMESPACE_PREFIXES", "")
	t.Setenv("SYSTEM_NAMESPACES", "")

	_, err := Load()
	if err == nil {
		t.Fatal("expected error for invalid METRICS_PORT, got nil")
	}
}

func TestLoad_OutOfRangeMetricsPort(t *testing.T) {
	t.Setenv("CHECK_INTERVAL", "")
	t.Setenv("METRICS_PORT", "99999")
	t.Setenv("SYSTEM_NAMESPACE_PREFIXES", "")
	t.Setenv("SYSTEM_NAMESPACES", "")

	_, err := Load()
	if err == nil {
		t.Fatal("expected error for METRICS_PORT=99999, got nil")
	}
}

func TestLoad_CustomNamespacePrefixes(t *testing.T) {
	t.Setenv("CHECK_INTERVAL", "")
	t.Setenv("METRICS_PORT", "")
	t.Setenv("SYSTEM_NAMESPACE_PREFIXES", "openshift-,my-infra-")
	t.Setenv("SYSTEM_NAMESPACES", "")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if len(cfg.SystemNamespacePrefixes) != 2 {
		t.Fatalf("expected 2 prefixes, got %d: %v", len(cfg.SystemNamespacePrefixes), cfg.SystemNamespacePrefixes)
	}
	if cfg.SystemNamespacePrefixes[0] != "openshift-" {
		t.Errorf("expected first prefix=openshift-, got %q", cfg.SystemNamespacePrefixes[0])
	}
	if cfg.SystemNamespacePrefixes[1] != "my-infra-" {
		t.Errorf("expected second prefix=my-infra-, got %q", cfg.SystemNamespacePrefixes[1])
	}
}

func TestLoad_CustomSystemNamespaces(t *testing.T) {
	t.Setenv("CHECK_INTERVAL", "")
	t.Setenv("METRICS_PORT", "")
	t.Setenv("SYSTEM_NAMESPACE_PREFIXES", "")
	t.Setenv("SYSTEM_NAMESPACES", "kube-system,kube-public,kube-node-lease,monitoring")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if len(cfg.SystemNamespaces) != 4 {
		t.Fatalf("expected 4 namespaces, got %d: %v", len(cfg.SystemNamespaces), cfg.SystemNamespaces)
	}
	found := false
	for _, ns := range cfg.SystemNamespaces {
		if ns == "monitoring" {
			found = true
		}
	}
	if !found {
		t.Error("expected 'monitoring' in SystemNamespaces")
	}
}
