package boundaries

import (
	"os"
	"path/filepath"
	"sort"
	"testing"

	"go-api/internal/modulemeta"
)

func TestEveryPluginModuleHasPluginSpec(t *testing.T) {
	var missing []string
	for name := range modulemeta.PluginModules {
		if _, ok := modulemeta.PluginSpecs[name]; !ok {
			missing = append(missing, name)
		}
	}
	sort.Strings(missing)
	if len(missing) > 0 {
		t.Fatalf("plugin modules missing plugin specs: %v", missing)
	}
}

func TestPluginSpecsMatchPluginModules(t *testing.T) {
	var extras []string
	for name := range modulemeta.PluginSpecs {
		if !modulemeta.PluginModules[name] {
			extras = append(extras, name)
		}
	}
	sort.Strings(extras)
	if len(extras) > 0 {
		t.Fatalf("plugin specs found for non-plugin modules: %v", extras)
	}
}

func TestPluginSpecsDeclareRouteSurfaceCorrectly(t *testing.T) {
	for name, spec := range modulemeta.PluginSpecs {
		routePath := filepath.Join("..", "modules", name, "routes.go")
		_, err := os.Stat(routePath)
		hasRoutes := err == nil
		if hasRoutes != spec.HasRoutes {
			t.Fatalf("plugin %s route declaration mismatch: spec=%v actual=%v", name, spec.HasRoutes, hasRoutes)
		}
	}
}

func TestPluginSpecsDeclareStatus(t *testing.T) {
	for name, spec := range modulemeta.PluginSpecs {
		if spec.Status == "" {
			t.Fatalf("plugin %s missing status", name)
		}
		switch spec.Status {
		case "active", "compat_retained":
		default:
			t.Fatalf("plugin %s has invalid status %q", name, spec.Status)
		}
	}
}

func TestPluginSpecsDeclareGovernanceFields(t *testing.T) {
	for name, spec := range modulemeta.PluginSpecs {
		if spec.RuntimeOwner == "" {
			t.Fatalf("plugin %s missing runtime owner", name)
		}
		switch spec.RuntimeOwner {
		case "module", "cron_bridge":
		default:
			t.Fatalf("plugin %s has invalid runtime owner %q", name, spec.RuntimeOwner)
		}
		if spec.HasPluginRuntime && spec.RuntimeOwner == "module" && contains(modulemeta.CronBridgePlugins, name) {
			t.Fatalf("plugin %s cannot be both module-owned and cron-bridge listed", name)
		}
		if spec.RetirementRule == "" {
			t.Fatalf("plugin %s missing retirement rule", name)
		}
		if spec.HasPluginRuntime && spec.RuntimeEntry == "" {
			t.Fatalf("plugin %s missing runtime entry", name)
		}
		if !spec.HasPluginRuntime && spec.RuntimeEntry != "" {
			t.Fatalf("plugin %s should not declare runtime entry without plugin runtime", name)
		}
	}
}

func contains(items []string, target string) bool {
	for _, item := range items {
		if item == target {
			return true
		}
	}
	return false
}
