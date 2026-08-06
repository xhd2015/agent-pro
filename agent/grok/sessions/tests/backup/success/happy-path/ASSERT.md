## Expected

- `Backup` succeeds with non-nil `BackupResult`.
- `Result.Dir` is a non-empty existing directory containing `manifest.json` and `payload/`.
- `Result.SessionID` / `CWD` / `CWDKey` match the parent fixture.
- Payload includes recursive parent session dir (marker file present).
- Payload includes child session dir (child marker present).
- Payload includes `sessions/<cwd_key>/prompt_history.session.jsonl` with parent+child lines only
  (noise session id absent).
- Payload includes `bookkeeping/relocations/<parent-id>.lock`.
- No `session_search.sqlite` under payload; no `unified.jsonl` under payload.
- Manifest: `version=1`, `kind=agent-pro.grok.session.backup`, identity fields,
  `related_sessions` includes parent and child, `check_results` all `ok=true`
  when present.

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

	assertEqualString(t, "CWD", resp.Result.CWD, req.CWD)
	assertEqualString(t, "CWDKey", resp.Result.CWDKey, req.CWDKey)

	parentPay := payloadSessionDir(dir, req.CWDKey, req.SessionID)
	childPay := payloadSessionDir(dir, req.CWDKey, req.ChildSessionID)
	assertDirExists(t, parentPay)
	assertDirExists(t, childPay)
	assertFileExists(t, filepath.Join(parentPay, "summary.json"))
	assertFileEqualsMarker(t, filepath.Join(parentPay, "marker.txt"), req.ParentMarker)
	assertFileEqualsMarker(t, filepath.Join(childPay, "marker.txt"), req.ChildMarker)

	ph := filepath.Join(dir, "payload", "sessions", req.CWDKey, "prompt_history.session.jsonl")
	assertFileExists(t, ph)
	body := readFileString(t, ph)
	assertContains(t, body, req.SessionID)
	assertContains(t, body, req.ChildSessionID)
	assertNotContains(t, body, req.PromptNoiseID)

	lockPay := filepath.Join(dir, "payload", "bookkeeping", "relocations", req.SessionID+".lock")
	assertFileExists(t, lockPay)

	if walkHasSuffix(t, filepath.Join(dir, "payload"), "session_search.sqlite") {
		t.Fatal("payload must not contain session_search.sqlite")
	}
	if walkHasSuffix(t, filepath.Join(dir, "payload"), "unified.jsonl") {
		t.Fatal("payload must not contain unified.jsonl log copy")
	}

	man := loadManifest(t, dir)
	assertManifestCore(t, man, req)

	related := asStringSlice(t, man["related_sessions"])
	if !sliceContains(related, req.SessionID) || !sliceContains(related, req.ChildSessionID) {
		t.Fatalf("related_sessions = %v, want parent+child", related)
	}

	// All recorded check_results should be ok when present.
	if cr, ok := man["check_results"].(map[string]any); ok {
		for name, raw := range cr {
			m, _ := raw.(map[string]any)
			if m == nil {
				t.Fatalf("check_results[%s] not an object: %v", name, raw)
			}
			okVal, _ := m["ok"].(bool)
			if !okVal {
				t.Fatalf("check_results[%s].ok = false detail=%v", name, m["detail"])
			}
		}
	}
}
```
