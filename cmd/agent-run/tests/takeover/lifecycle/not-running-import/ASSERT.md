## Expected

- Exit code 0.
- Stdout summary mentions session-id / provider / opened new iTerm2 window
  (flexible wording).
- At least one agent-run session meta exists with `runner_session_id` equal to
  the provider UUID and runner `grok-tty` (or grok-tty family).
- iTerm script is ModeForceNew (`create window`, no `create tab`) and follow-up
  mentions agent-run / open or resume / provider or session id.
- No kill log entries.

## Exit Code

0

```go
import (
	"encoding/json"
	"os"
	"path/filepath"
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

	assertContainsAny(t, combined,
		"session-id",
		"session id",
		"opened",
		"iterm",
		"iTerm",
	)
	assertContainsAny(t, combined, takeoverFixtureSessionID, "provider", "grok")

	// Meta: import must create a mapping for the provider UUID.
	ids := listAgentSessionIDs(t, req.Home)
	if len(ids) == 0 {
		t.Fatalf("expected agent-run meta after import; sessions dir empty")
	}
	foundMap := false
	for _, id := range ids {
		b, rerr := os.ReadFile(metaJSONPath(req.Home, id))
		if rerr != nil {
			t.Fatalf("read meta %s: %v", id, rerr)
		}
		var meta map[string]any
		if jerr := json.Unmarshal(b, &meta); jerr != nil {
			t.Fatalf("meta json %s: %v\n%s", id, jerr, b)
		}
		rsid, _ := meta["runner_session_id"].(string)
		runner, _ := meta["runner"].(string)
		if rsid == takeoverFixtureSessionID {
			foundMap = true
			if runner != "" && runner != "grok-tty" && runner != "grok" {
				t.Fatalf("import meta runner=%q want grok-tty/grok; meta=%s", runner, b)
			}
			break
		}
	}
	if !foundMap {
		t.Fatalf("no meta with runner_session_id=%s under %s; ids=%v",
			takeoverFixtureSessionID, filepath.Join(req.Home, "sessions"), ids)
	}

	script := readItermScript(t, req)
	assertItermForceNewScript(t, script)
	assertContainsAny(t, script, "agent-run", "resume", "open", "--resume-from-grok-session", takeoverFixtureSessionID)

	assertNoKillLog(t, req)
	_ = strings.ToLower(combined)
}
```
