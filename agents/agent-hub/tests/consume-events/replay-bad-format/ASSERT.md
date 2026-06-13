## Expected
- ExitCode != 0, error about format.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
    if resp.ExitCode == 0 {
        t.Fatal("expected non-zero exit code")
    }
    assertContains(t, resp.Stderr, "format")
}
```
