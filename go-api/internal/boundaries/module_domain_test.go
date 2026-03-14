package boundaries

import (
	"os"
	"path/filepath"
	"sort"
	"testing"

	"go-api/internal/modulemeta"
)

func TestAllModulesAreClassifiedIntoDomains(t *testing.T) {
	root := filepath.Clean(filepath.Join("..", "modules"))
	entries, err := os.ReadDir(root)
	if err != nil {
		t.Fatalf("read modules dir: %v", err)
	}

	classified := make(map[string]string)
	for name := range modulemeta.CoreModules {
		classified[name] = "core"
	}
	for name := range modulemeta.PluginModules {
		if prev, ok := classified[name]; ok {
			t.Fatalf("module %s classified twice: %s and plugin", name, prev)
		}
		classified[name] = "plugin"
	}
	for name := range modulemeta.SharedModules {
		if prev, ok := classified[name]; ok {
			t.Fatalf("module %s classified twice: %s and shared", name, prev)
		}
		classified[name] = "shared"
	}

	var unclassified []string
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		name := entry.Name()
		if _, ok := classified[name]; !ok {
			unclassified = append(unclassified, name)
		}
	}
	sort.Strings(unclassified)
	if len(unclassified) > 0 {
		t.Fatalf("unclassified modules found: %v", unclassified)
	}
}

func TestCoreAndPluginModuleSetsDoNotOverlap(t *testing.T) {
	for name := range modulemeta.CoreModules {
		if modulemeta.PluginModules[name] {
			t.Fatalf("module %s cannot be both core and plugin", name)
		}
	}
}
