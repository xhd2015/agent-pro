## Expected

- No API error.
- `BinaryLookedUp` is exactly `agent-run`.
- At least one launch call.

## Errors

- None.

## Exit Code

N/A

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertNoError(t, err)
	assertNoAPIError(t, resp)
	assertEqual(t, "BinaryLookedUp", resp.BinaryLookedUp, "agent-run")
	if resp.LaunchCalls < 1 {
		t.Fatal("expected launch call")
	}
}
```
