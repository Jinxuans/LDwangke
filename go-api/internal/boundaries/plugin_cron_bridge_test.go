package boundaries

import (
	"testing"

	"go-api/internal/modulemeta"
)

func TestCronBridgePluginsStayExplicit(t *testing.T) {
	if len(modulemeta.CronBridgePlugins) == 0 {
		t.Fatal("expected explicit cron bridge plugin list")
	}

	seen := map[string]bool{}
	for _, name := range modulemeta.CronBridgePlugins {
		if seen[name] {
			t.Fatalf("duplicate cron bridge plugin: %s", name)
		}
		seen[name] = true

		spec, ok := modulemeta.PluginSpecs[name]
		if !ok {
			t.Fatalf("cron bridge plugin %s missing PluginSpec", name)
		}
		if spec.Status != "active" {
			t.Fatalf("cron bridge plugin %s must stay active, got %s", name, spec.Status)
		}
		if !spec.HasPluginRuntime {
			t.Fatalf("cron bridge plugin %s must declare HasPluginRuntime", name)
		}
		if spec.RuntimeOwner != "cron_bridge" {
			t.Fatalf("cron bridge plugin %s must declare runtime owner cron_bridge, got %s", name, spec.RuntimeOwner)
		}
		if spec.RetirementRule == "" {
			t.Fatalf("cron bridge plugin %s must declare a retirement rule", name)
		}
	}
}
