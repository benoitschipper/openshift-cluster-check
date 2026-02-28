// Package checker implements the health check logic for OpenShift platform components.
package checker

import (
	"strings"

	"github.com/openshift-cluster-check/health-checker/internal/config"
)

// IsSystemNamespace returns true if the given namespace name is considered a
// system/platform namespace based on the configured prefix and exact-match rules.
//
// A namespace is a system namespace if:
//   - Its name starts with any prefix in cfg.SystemNamespacePrefixes, OR
//   - Its name exactly matches any entry in cfg.SystemNamespaces.
func IsSystemNamespace(name string, cfg config.Config) bool {
	// Check prefix rules
	for _, prefix := range cfg.SystemNamespacePrefixes {
		if strings.HasPrefix(name, prefix) {
			return true
		}
	}
	// Check exact-match rules
	for _, ns := range cfg.SystemNamespaces {
		if name == ns {
			return true
		}
	}
	return false
}
