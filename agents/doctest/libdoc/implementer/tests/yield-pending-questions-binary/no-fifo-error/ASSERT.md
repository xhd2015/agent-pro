## Expected
- Exit code non-zero.
- Stderr contains "QUESTION_FIFO must be set".

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
    if resp.ExitCode == 0 {
        t.Fatal("expected non-zero exit when QUESTION_FIFO is not set")
    }
    assertContains(t, resp.Stderr, "QUESTION_FIFO must be set")
}
```
