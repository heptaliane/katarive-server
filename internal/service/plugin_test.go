package service_test

import (
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/hashicorp/go-hclog"
	"github.com/heptaliane/katarive-server/internal/service"
)

func TestPluginRegistry(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	cases := map[string]struct {
		path                 string
		expectedSourceSize   int
		expectedNarratorSize int
		isError              bool
	}{
		"source": {
			path:                 buildPlugin(t, "source_plugin", tmpDir),
			expectedSourceSize:   1,
			expectedNarratorSize: 0,
			isError:              false,
		},
		"narrator": {
			path:                 buildPlugin(t, "narrator_plugin", tmpDir),
			expectedSourceSize:   0,
			expectedNarratorSize: 1,
			isError:              false,
		},
		"invalid": {
			path:                 buildPlugin(t, "invalid_plugin", tmpDir),
			expectedSourceSize:   0,
			expectedNarratorSize: 0,
			isError:              true,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			pr := service.NewPluginRegistry(hclog.Warn)

			err := pr.Load(tc.path)
			if err != nil {
				if !tc.isError {
					t.Errorf("Unexpected Error: %v", err)
					return
				}
			} else {
				if tc.isError {
					t.Errorf("Error expceted but got nil")
					return
				}
			}

			sources := pr.GetSources()
			narrators := pr.GetNarrators()

			if len(sources) != tc.expectedSourceSize {
				t.Errorf(
					"Source size unmatch: expect %d but got %d",
					tc.expectedSourceSize,
					len(sources),
				)
				return
			}
			if len(narrators) != tc.expectedNarratorSize {
				t.Errorf(
					"Narrator size unmatch: expect %d but got %d",
					tc.expectedNarratorSize,
					len(narrators),
				)
				return
			}
		})
	}
}

// Test helper
func buildPlugin(t *testing.T, testdataName string, destDir string) string {
	t.Helper()

	binName := testdataName
	if runtime.GOOS == "windows" {
		binName += ".exe"
	}
	destPath := filepath.Join(destDir, binName)
	srcPath := filepath.Join("testdata", "plugin", testdataName+".go")

	cmd := exec.Command("go", "build", "-o", destPath, srcPath)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to build test plugin %s: %v\n%s", testdataName, err, string(out))
	}

	return destPath
}
