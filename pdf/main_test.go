package pdf

import (
	"os"
	"path/filepath"
	"testing"
)

func findModuleRoot() string {
	dir, err := os.Getwd()
	if err != nil {
		return ""
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return ""
		}
		dir = parent
	}
}

func TestMain(m *testing.M) {
	if root := findModuleRoot(); root != "" {
		_ = os.Chdir(root)
	}
	os.Exit(m.Run())
}
