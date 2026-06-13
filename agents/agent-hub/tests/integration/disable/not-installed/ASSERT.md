## Expected
- Exit code non-zero.
- Error indicates plugin not installed.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
    if resp.ExitCode == 0 {
        t.Fatal("expected non-zero exit code when disabling absent plugin")
    }
}
```
