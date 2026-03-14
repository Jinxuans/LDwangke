package boundaries

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestServiceImportsStayWithinApprovedBoundaries(t *testing.T) {
	root := filepath.Clean(filepath.Join(".."))
	allowed := map[string]bool{
		"pluginruntime/cron_bridge.go": true,
	}

	var offenders []string
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}

		normalized := filepath.ToSlash(path)
		normalized = strings.TrimPrefix(normalized, "../")
		if strings.HasPrefix(normalized, "service/") {
			return nil
		}

		content, readErr := os.ReadFile(path)
		if readErr != nil {
			return readErr
		}
		if strings.Contains(string(content), `"go-api/internal/service"`) && !allowed[normalized] {
			offenders = append(offenders, normalized)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("scan internal files: %v", err)
	}
	if len(offenders) > 0 {
		t.Fatalf("service imports found outside approved compatibility boundaries: %v", offenders)
	}
}
