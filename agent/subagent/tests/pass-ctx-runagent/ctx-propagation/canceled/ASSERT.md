## Expected
- `runAgent` returns a non-nil error when called with a pre-canceled context.
- The error originates from the context cancellation propagating through `runner.Agent.Ask`.

## Exit Code
- Non-zero exit, errors are expected.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
    if err == nil {
        t.Fatalf("expected error from canceled context, got nil (output=%q)", resp.Output)
    }
}
```
