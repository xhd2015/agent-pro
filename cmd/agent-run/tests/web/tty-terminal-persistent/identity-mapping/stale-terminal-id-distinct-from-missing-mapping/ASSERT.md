## Expected

- HTTP 200.
- JSON reports `available:false`.
- JSON still includes `terminal_session_id:"session-1"` so callers can distinguish
  stale PTY from a session with no terminal mapping.

## Side Effects

- None beyond isolated fixture files.

## Errors

- None from `Run`.

## Exit Code

- Test process exits non-zero until unavailable status preserves mapped terminal id.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	assertHTTPStatus(t, resp, 200)
	requireTerminalMappingUnavailable(t, req, resp.HTTPBody)
}
```
