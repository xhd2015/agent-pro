## Expected

- Backup succeeds.
- `Result.Dir` equals the requested `OutDir` (cleaned).
- `OutDir/manifest.json` and `OutDir/payload/` exist with parent session payload.

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
	if filepath.Clean(dir) != filepath.Clean(req.OutDir) {
		t.Fatalf("Result.Dir = %q, want OutDir %q", dir, req.OutDir)
	}
	assertDirExists(t, payloadSessionDir(dir, req.CWDKey, req.SessionID))
	man := loadManifest(t, dir)
	assertManifestCore(t, man, req)
}
```
