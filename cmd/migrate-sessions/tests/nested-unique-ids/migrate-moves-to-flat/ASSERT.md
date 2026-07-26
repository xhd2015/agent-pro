## Expected

- Exit code 0.
- Flat dirs `sessions/sess_a` and `sessions/sess_b` exist with meta + events.
- Nested runner dirs gone (or empty removed).
- `sessions/.layout` has version 2.
- Backup directory exists under `home/backups/sessions-*`.
- `fake-codex-registry/live/marker.txt` still present.
- meta.runner preserved for each session.

```go
import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertExitZero(t, resp)

	assertDirExists(t, filepath.Join(req.Home, "sessions", "sess_a"))
	assertDirExists(t, filepath.Join(req.Home, "sessions", "sess_b"))
	assertFileExists(t, filepath.Join(req.Home, "sessions", "sess_a", "meta.json"))
	assertFileExists(t, filepath.Join(req.Home, "sessions", "sess_b", "meta.json"))
	assertFileExists(t, filepath.Join(req.Home, "sessions", "sess_a", "events.jsonl"))
	assertFileExists(t, filepath.Join(req.Home, "sessions", "sess_b", "events.jsonl"))

	// nested paths should not remain as session homes
	assertPathMissing(t, filepath.Join(req.Home, "sessions", "fake-codex", "sess_a"))
	assertPathMissing(t, filepath.Join(req.Home, "sessions", "fake-opencode", "sess_b"))

	if v := layoutVersion(t, req.Home); v != 2 {
		t.Fatalf(".layout version = %d, want 2", v)
	}

	backs := backupDirsUnder(t, req.Home)
	if len(backs) == 0 {
		t.Fatalf("expected backup under %s/backups/sessions-*, none found", req.Home)
	}
	// backup should contain nested snapshot
	foundBackupMeta := false
	_ = filepath.Walk(backs[0], func(path string, info os.FileInfo, err error) error {
		if err == nil && strings.HasSuffix(path, "meta.json") {
			foundBackupMeta = true
		}
		return nil
	})
	if !foundBackupMeta {
		t.Fatalf("backup %s has no meta.json", backs[0])
	}

	assertFileExists(t, filepath.Join(req.Home, "fake-codex-registry", "live", "marker.txt"))

	if got := readMetaRunner(t, filepath.Join(req.Home, "sessions", "sess_a", "meta.json")); got != "fake-codex" {
		t.Fatalf("sess_a runner = %q want fake-codex", got)
	}
	if got := readMetaRunner(t, filepath.Join(req.Home, "sessions", "sess_b", "meta.json")); got != "fake-opencode" {
		t.Fatalf("sess_b runner = %q want fake-opencode", got)
	}

	// events content preserved
	ea, _ := os.ReadFile(filepath.Join(req.Home, "sessions", "sess_a", "events.jsonl"))
	eb, _ := os.ReadFile(filepath.Join(req.Home, "sessions", "sess_b", "events.jsonl"))
	if !strings.Contains(string(ea), "event-a") {
		t.Fatalf("sess_a events missing event-a: %s", ea)
	}
	if !strings.Contains(string(eb), "event-b") {
		t.Fatalf("sess_b events missing event-b: %s", eb)
	}
}
```
