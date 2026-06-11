## Expected
- Exit code non-zero.
- Stderr contains "requires".

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
    if resp.ExitCode == 0 {
        t.Fatal("expected non-zero exit for missing prompt")
    }
    assertContains(t, resp.Stderr, "requires")
}
```
