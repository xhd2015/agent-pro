## Preconditions
- This branch tests real tool execution during `fake-opencode run`.
- The mock config must include `"runner":"fake-opencode"` and at least one stdout event.

## Steps
1. Use `writeMockConfig` and the existing `Run` harness from the parent SETUP.md.
2. This branch provides a `createTestFile` helper that writes a file in a known location
   for read, write, and grep tests.

```go
import (
    "os"
    "path/filepath"
    "testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
    req.Env = append(req.Env, "FAKE_OPENCODE_TEST_MODE=tool-exec")
    return nil
}

func createTestFile(t *testing.T, req *Request, relPath string, content string) string {
    t.Helper()
    absPath := filepath.Join(req.TempDir, relPath)
    dir := filepath.Dir(absPath)
    if err := os.MkdirAll(dir, 0755); err != nil {
        t.Fatalf("mkdir %s: %v", dir, err)
    }
    if err := os.WriteFile(absPath, []byte(content), 0644); err != nil {
        t.Fatalf("write %s: %v", absPath, err)
    }
    return absPath
}
```
