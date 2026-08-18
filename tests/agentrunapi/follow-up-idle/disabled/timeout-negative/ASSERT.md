## Expected

- `NormalizeIdle(false, -1s)` succeeds (timeout ignored; no error).
- No emit error.
- Follow-up line omits both idle-exit flags.

## Side Effects

- None (pure).

## Errors

- None.

## Exit Code

N/A

```go
import "testing"

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	t.Helper()
	_ = d
	_ = req
	assertNoError(t, err)
	assertNoAPIError(t, resp)
	if resp.Enabled {
		t.Fatal("NormalizeIdle(false, -1s) must return enabled=false")
	}
	assertOmitsIdleFlags(t, resp.FollowUp)
	assertOpenProfile(t, resp.FollowUp, "sess-idle-off-neg")
}
```
