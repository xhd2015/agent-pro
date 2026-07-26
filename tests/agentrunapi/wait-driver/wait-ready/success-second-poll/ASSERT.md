## Expected

- No API error.
- At least two StatusFn polls (not-ready then ready).

## Side Effects

- StatusFn only; no agent-run binary.

## Errors

- None.

## Exit Code

N/A

```go
import "testing"

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertNoError(t, err)
	assertNoAPIError(t, resp)
	if resp.StatusPollCalls < 2 {
		t.Fatalf("StatusPollCalls: got %d, want >= 2", resp.StatusPollCalls)
	}
}
```
