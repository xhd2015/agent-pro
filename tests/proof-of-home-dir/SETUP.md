## Preconditions
- `bash`, `go`, `node`, and `bun` are available in PATH (individual tests skip if not present).
- No code under test from this project — these are pure platform behavior demonstrations.

## Steps
1. Create a temporary directory to serve as the fake `HOME`.
2. Run the target runtime with `HOME` set to the temporary directory.
3. Capture the runtime's reported home directory.
4. Assert it matches the temporary `HOME` directory.

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

func Setup(t *testing.T, req *Request) error {
    req.TmpHome = filepath.Join(t.TempDir(), "fake-home")
    if err := os.MkdirAll(req.TmpHome, 0755); err != nil {
        return fmt.Errorf("create fake HOME: %w", err)
    }
    return nil
}

// assertHomeDir is a helper to check that the runtime's reported home
// matches the temporary HOME directory.
func assertHomeDir(t *testing.T, req *Request, resp *Response, err error) {
    t.Helper()
    if err != nil && resp.ExitCode == 0 {
        t.Fatalf("run failed: %v", err)
    }
    if resp.ExitCode != 0 {
        t.Fatalf("exit code = %d, stderr:\n%s", resp.ExitCode, resp.Stderr)
    }
    got := strings.TrimSpace(resp.Stdout)
    if got != req.TmpHome {
        t.Fatalf("expected home %q, got %q", req.TmpHome, got)
    }
}
```
