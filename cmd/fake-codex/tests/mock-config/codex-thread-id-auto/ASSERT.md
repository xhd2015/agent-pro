## Expected
- Exit code 0.
- The marker file contains a non-empty CODEX_THREAD_ID starting with `codex_`.

```go
import (
    "os"
    "path/filepath"
    "strings"
    "testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
    assertSuccess(t, resp)

    marker := filepath.Join(req.TempDir, "thread-id.txt")
    data, readErr := os.ReadFile(marker)
    if readErr != nil {
        t.Fatalf("read marker file: %v", readErr)
    }
    got := strings.TrimSpace(string(data))
    if got == "" {
        t.Fatal("CODEX_THREAD_ID is empty, expected auto-generated ID")
    }
    if !strings.HasPrefix(got, "codex_") {
        t.Fatalf("CODEX_THREAD_ID = %q, expected prefix %q", got, "codex_")
    }
}
```
