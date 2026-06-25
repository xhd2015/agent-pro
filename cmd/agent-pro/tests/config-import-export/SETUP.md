## Preconditions
- The repository contains `pkgs/agentconfig` package.
- Each test runs in a temporary directory with a synthetic home directory.
- No real agent configuration is used — all source/destination paths are under `HomeDir`.

## Steps
1. Create a temporary home directory.
2. Leaf SETUP.md creates any source config files (for export) or pre-populated zip files (for import).
3. Run calls `agentconfig.Export()` or `agentconfig.Import()` depending on Operation.
4. Leaf ASSERT.md inspects the zip file (export) or files on disk (import).

## Context
- Helper functions are provided for creating source files, creating zip files, reading zip entries, and checking file permissions.
- The `HomeDir` is used as the user's home directory root for resolving all source/destination paths.

```go
import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	agentconfig "github.com/xhd2015/agent-pro/pkgs/agentconfig"
)

func Setup(t *testing.T, req *Request) error {
	req.HomeDir = t.TempDir()
	req.ZipPath = filepath.Join(req.HomeDir, "config.zip")
	return nil
}

func createSourceFile(t *testing.T, homeDir string, relPath string, content string) {
	t.Helper()
	absPath := filepath.Join(homeDir, relPath)
	if err := os.MkdirAll(filepath.Dir(absPath), 0755); err != nil {
		t.Fatalf("create source dir: %v", err)
	}
	if err := os.WriteFile(absPath, []byte(content), 0644); err != nil {
		t.Fatalf("create source file: %v", err)
	}
}

func createZip(t *testing.T, zipPath string, entries map[string]string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(zipPath), 0755); err != nil {
		t.Fatalf("create zip dir: %v", err)
	}
	fw, err := os.Create(zipPath)
	if err != nil {
		t.Fatalf("create zip: %v", err)
	}
	defer fw.Close()
	zw := zip.NewWriter(fw)
	defer zw.Close()
	for name, content := range entries {
		w, err := zw.Create(name)
		if err != nil {
			t.Fatalf("create zip entry %s: %v", name, err)
		}
		if _, err := w.Write([]byte(content)); err != nil {
			t.Fatalf("write zip entry %s: %v", name, err)
		}
	}
}

func readZipEntries(t *testing.T, zipPath string) []string {
	t.Helper()
	if _, err := os.Stat(zipPath); os.IsNotExist(err) {
		return nil
	}
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		t.Fatalf("open zip: %v", err)
	}
	defer r.Close()
	var names []string
	for _, f := range r.File {
		names = append(names, f.Name)
	}
	return names
}

func assertFileExists(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatalf("expected file %s to exist", path)
	}
}

func assertFileNotExist(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); err == nil {
		t.Fatalf("expected file %s to not exist", path)
	}
}

func assertFileMode(t *testing.T, path string, want os.FileMode) {
	t.Helper()
	fi, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat %s: %v", path, err)
	}
	got := fi.Mode() & os.ModePerm
	if got != want {
		t.Fatalf("file %s mode = %o, want %o", path, got, want)
	}
}

func assertFileContent(t *testing.T, path string, want string) {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	if string(data) != want {
		t.Fatalf("file %s content = %q, want %q", path, string(data), want)
	}
}

func assertZipContains(t *testing.T, zipPath string, wantNames []string) {
	t.Helper()
	got := readZipEntries(t, zipPath)
	if len(got) != len(wantNames) {
		t.Fatalf("zip has %d entries, want %d\ngot:  %v\nwant: %v", len(got), len(wantNames), got, wantNames)
	}
	for _, want := range wantNames {
		found := false
		for _, g := range got {
			if g == want {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("zip missing entry %q; has: %v", want, got)
		}
	}
}

func assertSuccess(t *testing.T, resp *Response) {
	t.Helper()
	if resp.Err != nil {
		t.Fatalf("operation failed: %v", resp.Err)
	}
}

func assertError(t *testing.T, resp *Response) {
	t.Helper()
	if resp.Err == nil {
		t.Fatal("expected error but got nil")
	}
}
```
