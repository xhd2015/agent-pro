## Expected

- `Backup` succeeds (no error) with non-nil `BackupResult`.
- `Result.DryRun == true`.
- `Result.Dir`, `ArchivePath`, `ManifestPath` are empty.
- `Result.PlannedFiles > 0`; `PlannedBytes >= 0`.
- `Result.RelatedSessions` includes parent and child session ids.
- `Result.SessionID` / `CWD` / `CWDKey` match fixture identity.
- Requested `OutDir` does **not** exist after the call (not created).
- No `manifest.json` / `payload/` written under `OutDir`.

## Errors

- None.

```go
import (
	"testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertNoHarnessErr(t, err)
	assertDryRunSuccess(t, req, resp)

	assertEqualString(t, "CWD", resp.Result.CWD, req.CWD)
	assertEqualString(t, "CWDKey", resp.Result.CWDKey, req.CWDKey)

	related := resp.Result.RelatedSessions
	if !sliceContains(related, req.SessionID) {
		t.Fatalf("RelatedSessions missing parent %q: %v", req.SessionID, related)
	}
	if !sliceContains(related, req.ChildSessionID) {
		t.Fatalf("RelatedSessions missing child %q: %v", req.ChildSessionID, related)
	}
	assertPathMissing(t, req.OutDir)
}
```
