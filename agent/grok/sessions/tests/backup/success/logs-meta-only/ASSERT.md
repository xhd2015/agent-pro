## Expected

- Backup succeeds.
- Manifest `logs.match_count` equals fixture match count (5).
- Manifest `logs.last_lines` length ≤ 3 and ≥ 1 when matches exist.
- Each last_line has non-empty `text`; `time` set when source line has parseable `ts`.
- No file named `unified.jsonl` (or logs tree) under `payload/`.

## Errors

- None.

```go
import (
	"os"
	"path/filepath"
	"testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertNoHarnessErr(t, err)
	dir := assertSuccessfulBackup(t, req, resp)

	if walkHasSuffix(t, filepath.Join(dir, "payload"), "unified.jsonl") {
		t.Fatal("payload must not contain unified.jsonl")
	}
	// Also forbid a payload/bookkeeping/logs tree with log body copies.
	if _, err := os.Stat(filepath.Join(dir, "payload", "bookkeeping", "logs")); err == nil {
		t.Fatal("payload/bookkeeping/logs must not exist (logs meta only)")
	}

	man := loadManifest(t, dir)
	logs, _ := man["logs"].(map[string]any)
	if logs == nil {
		t.Fatal("manifest.logs missing")
	}
	mc, ok := logs["match_count"].(float64)
	if !ok {
		t.Fatalf("logs.match_count missing/type: %v", logs["match_count"])
	}
	assertEqualInt(t, "logs.match_count", int(mc), req.LogMatchCount)

	rawLines, _ := logs["last_lines"].([]any)
	if len(rawLines) == 0 {
		t.Fatal("logs.last_lines empty despite matches")
	}
	if len(rawLines) > 3 {
		t.Fatalf("logs.last_lines len=%d, want ≤3", len(rawLines))
	}
	for i, raw := range rawLines {
		m, _ := raw.(map[string]any)
		if m == nil {
			t.Fatalf("last_lines[%d] not object: %v", i, raw)
		}
		text, _ := m["text"].(string)
		if text == "" {
			t.Fatalf("last_lines[%d].text empty", i)
		}
		// Standard fixture lines include "ts" — expect time when parseable.
		if tm, _ := m["time"].(string); tm == "" {
			// Allow empty only if implementer cannot parse; fixture is JSON with ts,
			// so require non-empty.
			t.Fatalf("last_lines[%d].time empty (fixture has ts)", i)
		}
	}
}
```
