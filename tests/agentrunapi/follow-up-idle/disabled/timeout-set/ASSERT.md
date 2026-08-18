## Expected

- `NormalizeIdle(false, 2s)` succeeds with enabled=false (timeout ignored; `d` unused).
- No emit error.
- Follow-up line has neither `--exit-on-idle` nor `--idle-timeout=2s` (nor any
  `--idle-timeout*` token).
- Open profile still present.

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
		t.Fatal("NormalizeIdle(false, 2s) must return enabled=false")
	}
	assertOmitsIdleFlags(t, resp.FollowUp)
	if hasExactToken(resp.FollowUp, "--idle-timeout=2s") {
		t.Fatalf("disabled emit must not include --idle-timeout=2s; got %q", resp.FollowUp)
	}
	assertOpenProfile(t, resp.FollowUp, "sess-idle-off-2s")
}
```
