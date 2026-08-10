## Expected

- Exit code 0.
- Reuses agent-run session id `takeover-mapped-resume-s1` (stdout/script/meta).
- Does **not** fail with already-mapped / already imported error.
- iTerm ForceNew script present; follow-up references existing session id and/or resume/open.
- No kill log.
- Exactly one mapped meta for the provider UUID (no duplicate import session).

## Exit Code

0

```go
import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	if err != nil {
		t.Fatal(err)
	}
	combined := combinedOut(resp)
	assertTakeoverActionImplemented(t, combined)
	assertExitCode(t, resp, 0)

	lower := strings.ToLower(combined)
	// Must not be the import-collision error path.
	if strings.Contains(lower, "already mapped") || strings.Contains(lower, "already imported") {
		t.Fatalf("mapped dead session should resume, not error as already-mapped:\n%s", combined)
	}

	agentID := "takeover-mapped-resume-s1"
	assertContainsAny(t, combined, agentID, "session-id", "resume", "opened", "iterm")

	// Meta still the single known mapping.
	if !metaExists(req.Home, agentID) {
		t.Fatalf("expected existing meta %s to remain", agentID)
	}
	b, rerr := os.ReadFile(metaJSONPath(req.Home, agentID))
	if rerr != nil {
		t.Fatalf("read meta: %v", rerr)
	}
	var meta map[string]any
	if jerr := json.Unmarshal(b, &meta); jerr != nil {
		t.Fatalf("meta json: %v", jerr)
	}
	if rsid, _ := meta["runner_session_id"].(string); rsid != takeoverFixtureSessionID {
		t.Fatalf("runner_session_id=%q want %s", rsid, takeoverFixtureSessionID)
	}

	// No extra sessions with same provider binding.
	mapped := 0
	for _, id := range listAgentSessionIDs(t, req.Home) {
		mb, _ := os.ReadFile(metaJSONPath(req.Home, id))
		var m map[string]any
		_ = json.Unmarshal(mb, &m)
		if rsid, _ := m["runner_session_id"].(string); rsid == takeoverFixtureSessionID {
			mapped++
		}
	}
	if mapped != 1 {
		t.Fatalf("want exactly 1 meta mapped to provider, got %d", mapped)
	}

	script := readItermScript(t, req)
	assertItermForceNewScript(t, script)
	assertContainsAny(t, script, agentID, "resume", "open", "agent-run")

	assertNoKillLog(t, req)
}
```
