## Expected
- The command exits with code 5 and emits the configured error.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
    if resp.ExitCode != 5 {
        t.Fatalf("exit code = %d, want 5; stderr=%s", resp.ExitCode, resp.Stderr)
    }
    assertContains(t, resp.Stdout, `"type":"error"`)
    assertContains(t, resp.Stderr, "planned opencode failure")
}
```

