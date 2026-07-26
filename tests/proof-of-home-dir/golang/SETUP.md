## Preconditions
- `go` is available in PATH (test is skipped otherwise).

## Steps
1. Write a minimal `main.go` and `go.mod` that calls `os.UserHomeDir()`.
2. Run the program from the fake `HOME` directory, with `HOME` overridden.
3. The output should be the temporary `HOME` directory path.

```go
import (
    "os"
    "os/exec"
    "path/filepath"
    "testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
    if _, err := exec.LookPath("go"); err != nil {
        t.Skipf("skipping: go not found in PATH")
    }

    projDir := filepath.Join(req.TmpHome, "go-hometest")
    if err := os.MkdirAll(projDir, 0755); err != nil {
        return err
    }

    goMod := "module hometest\n\ngo 1.21\n"
    if err := os.WriteFile(filepath.Join(projDir, "go.mod"), []byte(goMod), 0644); err != nil {
        return err
    }

    fixture := filepath.Join(d.DOCTEST_ROOT, "golang", "testdata", "main.go")
    src, err := os.ReadFile(fixture)
    if err != nil {
        return err
    }
    if err := os.WriteFile(filepath.Join(projDir, "main.go"), src, 0644); err != nil {
        return err
    }

    req.Cmd = "go"
    req.Args = []string{"run", "."}
    req.WorkDir = projDir
    return nil
}
```
