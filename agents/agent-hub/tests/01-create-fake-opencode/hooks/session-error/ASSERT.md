## Expected
- The command exits with code 6 and fires `session.error`.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
    if resp.ExitCode != 6 {
        t.Fatalf("exit code = %d, want 6; stderr=%s", resp.ExitCode, resp.Stderr)
    }
    assertContains(t, resp.HookLog, `"event":"session.error"`)
    assertContains(t, resp.HookLog, `"error":"bad session"`)
}
```

