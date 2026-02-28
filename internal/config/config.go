// Package config provides environment variable parsing and validation
// for the OpenShift cluster health-checker.
package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config holds all runtime configuration for the health-checker.
type Config struct {
	// CheckInterval is how often checks are run (default: 30s).
	CheckInterval time.Duration

	// MetricsPort is the HTTP port for the /metrics endpoint (default: 8080).
	MetricsPort int

	// SystemNamespacePrefixes is the list of namespace prefixes considered system namespaces.
	// Default: ["openshift-", "kube-"]
	// The "kube-" prefix covers kube-system, kube-public, kube-node-lease, and any
	// future kube-* namespaces added by Kubernetes, making the filter forward-compatible.
	SystemNamespacePrefixes []string

	// SystemNamespaces is the list of exact namespace names considered system namespaces.
	// Default: [] (empty — the "kube-" prefix in SystemNamespacePrefixes subsumes all
	// kube-* namespaces). Set this to add non-prefixed system namespaces (e.g., "monitoring").
	SystemNamespaces []string
}

// Load reads configuration from environment variables and applies defaults.
// Returns an error if any value fails validation.
func Load() (Config, error) {
	cfg := Config{}

	// CHECK_INTERVAL: positive integer seconds, default 30
	intervalStr := os.Getenv("CHECK_INTERVAL")
	if intervalStr == "" {
		cfg.CheckInterval = 30 * time.Second
	} else {
		secs, err := strconv.Atoi(intervalStr)
		if err != nil || secs <= 0 {
			return Config{}, fmt.Errorf("CHECK_INTERVAL must be a positive integer (got %q)", intervalStr)
		}
		cfg.CheckInterval = time.Duration(secs) * time.Second
	}

	// METRICS_PORT: valid port number 1-65535, default 8080
	portStr := os.Getenv("METRICS_PORT")
	if portStr == "" {
		cfg.MetricsPort = 8080
	} else {
		port, err := strconv.Atoi(portStr)
		if err != nil || port < 1 || port > 65535 {
			return Config{}, fmt.Errorf("METRICS_PORT must be a valid port number 1-65535 (got %q)", portStr)
		}
		cfg.MetricsPort = port
	}

	// SYSTEM_NAMESPACE_PREFIXES: comma-separated, default "openshift-,kube-"
	// The "kube-" prefix covers all current and future kube-* system namespaces.
	prefixStr := os.Getenv("SYSTEM_NAMESPACE_PREFIXES")
	if prefixStr == "" {
		cfg.SystemNamespacePrefixes = []string{"openshift-", "kube-"}
	} else {
		cfg.SystemNamespacePrefixes = splitAndTrim(prefixStr)
	}

	// SYSTEM_NAMESPACES: comma-separated, default "" (empty)
	// The "kube-" prefix in SYSTEM_NAMESPACE_PREFIXES subsumes all kube-* namespaces,
	// so no exact-match list is needed by default.
	nsStr := os.Getenv("SYSTEM_NAMESPACES")
	if nsStr == "" {
		cfg.SystemNamespaces = []string{}
	} else {
		cfg.SystemNamespaces = splitAndTrim(nsStr)
	}

	return cfg, nil
}

// splitAndTrim splits a comma-separated string and trims whitespace from each element.
func splitAndTrim(s string) []string {
	parts := strings.Split(s, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			result = append(result, p)
		}
	}
	return result
}
