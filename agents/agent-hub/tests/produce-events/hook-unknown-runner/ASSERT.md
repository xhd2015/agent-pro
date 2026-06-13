## Expected
- ExitCode != 0, stderr contains "unknown runner".

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
    if resp.ExitCode == 0 {
        t.Fatal("expected non-zero exit code")
    }
    assertContains(t, resp.Stderr, "unknown runner")
}
```
