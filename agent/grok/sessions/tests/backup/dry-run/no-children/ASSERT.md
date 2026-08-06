## Expected

- Dry-run succeeds.
- `Result.DryRun == true`; write paths empty; `PlannedFiles > 0`.
- `RelatedSessions` contains parent; does **not** contain child id.
- OutDir not created.

## Errors

- None.

```go
import "testing"

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertNoHarnessErr(t, err)
	assertDryRunSuccess(t, req, resp)

	related := resp.Result.RelatedSessions
	if !sliceContains(related, req.SessionID) {
		t.Fatalf("RelatedSessions missing parent: %v", related)
	}
	if sliceContains(related, req.ChildSessionID) {
		t.Fatalf("RelatedSessions should not include skipped child: %v", related)
	}
}
```
