## Expected
- The command exits with code 7.
- stdout and stderr preserve configured content.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
    if resp.ExitCode != 7 {
        t.Fatalf("exit code = %d, want 7; stderr=%s", resp.ExitCode, resp.Stderr)
    }
    assertContains(t, resp.Stdout, `"before failure"`)
    assertContains(t, resp.Stderr, "planned failure")
}
```

