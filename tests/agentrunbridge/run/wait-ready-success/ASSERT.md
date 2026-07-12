## Expected

- No API error.
- At least one launch call.
- At least two status poll calls (not-ready then ready).

## Errors

- None.

## Exit Code

N/A

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertNoError(t, err)
	assertNoAPIError(t, resp)
	if resp.LaunchCalls < 1 {
		t.Fatal("expected launch")
	}
	if resp.StatusPollCalls < 2 {
		t.Fatalf("expected >=2 status polls, got %d", resp.StatusPollCalls)
	}
}
```
