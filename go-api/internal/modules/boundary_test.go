package modules

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestModulesDoNotImportServicePackage(t *testing.T) {
	root := "."
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

		content, readErr := os.ReadFile(path)
		if readErr != nil {
			return readErr
		}
		if strings.Contains(string(content), `"go-api/internal/service"`) {
			offenders = append(offenders, filepath.ToSlash(path))
		}
		return nil
	})
	if err != nil {
		t.Fatalf("scan module files: %v", err)
	}
	if len(offenders) > 0 {
		t.Fatalf("module files must not import go-api/internal/service: %v", offenders)
	}
}
