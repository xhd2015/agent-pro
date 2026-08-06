## Expected

- Dry-run succeeds with plan fields set.
- `OutDir` path does not exist after call.
- `ArchivePath` path does not exist after call.
- Result write paths (`Dir`, `ArchivePath`, `ManifestPath`) empty.
- No payload/manifest under OutDir.

## Side Effects

- No directory or archive created for the planned output paths.

## Errors

- None.

```go
import (
	"os"
	"testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertNoHarnessErr(t, err)
	assertDryRunSuccess(t, req, resp)

	assertPathMissing(t, req.OutDir)
	if _, err := os.Stat(req.ArchivePath); err == nil {
		t.Fatalf("archive path must not exist after dry-run: %s", req.ArchivePath)
	} else if !os.IsNotExist(err) {
		t.Fatalf("stat archive path: %v", err)
	}
}
```
