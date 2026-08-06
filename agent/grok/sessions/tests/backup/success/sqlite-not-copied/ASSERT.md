## Expected

- Backup succeeds.
- Source sqlite still exists with original marker (untouched).
- No `session_search.sqlite` anywhere under `payload/`.
- Manifest `sqlite.present` is true; `sqlite.path` non-empty when provided.

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

	// Source untouched.
	got := readFileString(t, req.SQLitePath)
	if got != req.SQLiteMarker {
		t.Fatalf("source sqlite changed: %q", got)
	}

	if walkHasSuffix(t, filepath.Join(dir, "payload"), "session_search.sqlite") {
		t.Fatal("payload must not contain session_search.sqlite")
	}

	man := loadManifest(t, dir)
	sqlite, _ := man["sqlite"].(map[string]any)
	if sqlite == nil {
		t.Fatal("manifest.sqlite missing")
	}
	present, _ := sqlite["present"].(bool)
	if !present {
		t.Fatalf("sqlite.present = %v, want true", sqlite["present"])
	}
}
```
