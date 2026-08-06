## Expected

- Backup succeeds.
- Parent session dir present under payload.
- Child session dir **not** present under `payload/sessions/<cwd_key>/<child-id>/`.
- Parent recursive copy still includes `subagents/<child>/meta.json` (meta inside parent tree).
- `related_sessions` contains parent; child id should not be listed as a copied related session
  (or checks reflect no child payload — at minimum child dir absent).

## Errors

- None.

```go
import (
	"path/filepath"
	"testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertNoHarnessErr(t, err)
	dir := assertSuccessfulBackup(t, req, resp)

	parentPay := payloadSessionDir(dir, req.CWDKey, req.SessionID)
	childPay := payloadSessionDir(dir, req.CWDKey, req.ChildSessionID)
	assertDirExists(t, parentPay)
	assertPathMissing(t, childPay)

	// Parent tree still has subagent meta (recursive copy of parent).
	meta := filepath.Join(parentPay, "subagents", req.ChildSessionID, "meta.json")
	assertFileExists(t, meta)

	man := loadManifest(t, dir)
	assertManifestCore(t, man, req)
	related := asStringSlice(t, man["related_sessions"])
	if !sliceContains(related, req.SessionID) {
		t.Fatalf("related_sessions missing parent: %v", related)
	}
	// Child should not be treated as included related session when skipped.
	if sliceContains(related, req.ChildSessionID) {
		t.Fatalf("related_sessions should not include skipped child: %v", related)
	}
}
```
