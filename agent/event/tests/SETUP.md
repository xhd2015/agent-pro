## Preconditions
- The repository must have `go.mod` at its root.
- Each leaf provides a `main.go` source string that exercises the packages under test.

## Steps
1. Write `main.go` to the temporary directory.
2. Run `go run ./...` from the repo root referencing the temp dir main.go.
3. Capture stdout, stderr, and exit status.

```go
import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

type Request struct {
	RepoRoot string
	TempDir  string
	MainGo   string
}

type Response struct {
	ExitCode int
	Stdout   string
	Stderr   string
	Err      error
}

func Setup(t *testing.T, req *Request) error {
	_ = writeFile
	_ = assertSuccess
	_ = assertContains
	_ = assertNotContains
	req.RepoRoot = filepath.Clean(filepath.Join(DOCTEST_ROOT, "../../.."))
	if _, err := os.Stat(filepath.Join(req.RepoRoot, "go.mod")); err != nil {
		return fmt.Errorf("repo root not found: %w", err)
	}
	req.TempDir = t.TempDir()
	return nil
}

func Run(t *testing.T, req *Request) (*Response, error) {
	if req.MainGo == "" {
		return nil, fmt.Errorf("MainGo is required")
	}
	mainPath := filepath.Join(req.TempDir, "main.go")
	writeFile(t, mainPath, req.MainGo)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "go", "run", mainPath)
	cmd.Dir = req.RepoRoot
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	resp := &Response{
		Stdout: stdout.String(),
		Stderr: stderr.String(),
		Err:    err,
	}
	if err == nil {
		return resp, nil
	}
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		resp.ExitCode = exitErr.ExitCode()
		return resp, nil
	}
	if ctx.Err() != nil {
		return resp, ctx.Err()
	}
	return resp, err
}

func writeFile(t *testing.T, path string, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func assertSuccess(t *testing.T, resp *Response) {
	t.Helper()
	if resp.Err != nil && resp.ExitCode == 0 {
		t.Fatalf("run failed: %v\nstderr: %s", resp.Err, resp.Stderr)
	}
	if resp.ExitCode != 0 {
		t.Fatalf("exit code = %d, stderr:\n%s", resp.ExitCode, resp.Stderr)
	}
}

func assertContains(t *testing.T, got string, want string) {
	t.Helper()
	if !strings.Contains(got, want) {
		t.Fatalf("missing %q in:\n%s", want, got)
	}
}

func assertNotContains(t *testing.T, got string, want string) {
	t.Helper()
	if strings.Contains(got, want) {
		t.Fatalf("unexpected %q in:\n%s", want, got)
	}
}
```
