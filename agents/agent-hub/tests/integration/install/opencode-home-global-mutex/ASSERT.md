## Expected
- Exit code non-zero.
- Stderr indicates that `--opencode-home` and `--global` are mutually exclusive.

```go
import (
    "strings"
    "testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
    if resp.ExitCode == 0 {
        t.Fatal("expected non-zero exit code for mutually exclusive flags")
    }
    combined := resp.Stdout + resp.Stderr
    if !strings.Contains(strings.ToLower(combined), "exclusive") && !strings.Contains(strings.ToLower(combined), "cannot") {
        t.Fatalf("expected error about mutually exclusive flags, got:\nstdout:\n%s\nstderr:\n%s", resp.Stdout, resp.Stderr)
    }
}
```
