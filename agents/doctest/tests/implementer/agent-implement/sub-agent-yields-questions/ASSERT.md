## Expected
- Exit code 0 (yielding questions is not an error).
- Stdout contains the question JSON.

```go
import (
    "strings"
    "testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
    if err != nil {
        t.Fatalf("run failed: %v", err)
    }
    if resp.ExitCode != 0 {
        t.Fatalf("exit code = %d, want 0\nstderr:\n%s", resp.ExitCode, resp.Stderr)
    }
    if !strings.Contains(resp.Stdout, `"question"`) {
        t.Fatalf("stdout missing question:\n%s", resp.Stdout)
    }
    if !strings.Contains(resp.Stdout, `What is the target port?`) {
        t.Fatalf("stdout missing question text:\n%s", resp.Stdout)
    }
}
```
