## Expected

- Exit code 0 (auto-detect → codex-tty lifecycle import).
- Not the empty-runner error demanding `--grok`/`--codex`/`--agent-runner`.
- Agent-run meta maps provider UUID with runner `codex-tty` / `codex`.
- iTerm ForceNew script present; no kill log.

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
	lower := strings.ToLower(combined)
	if strings.Contains(lower, "requires --grok") || strings.Contains(lower, "requires --codex") {
		t.Fatalf("auto-detect should not demand runner flags when Codex has the id, got:\n%s", combined)
	}
	assertExitCode(t, resp, 0)

	assertContainsAny(t, combined,
		"session-id",
		"session id",
		"opened",
		"iterm",
		"iTerm",
	)
	assertContainsAny(t, combined, takeoverCodexFixtureSessionID, "provider", "codex")

	ids := listAgentSessionIDs(t, req.Home)
	if len(ids) == 0 {
		t.Fatalf("expected agent-run meta after auto-detect codex import; sessions empty")
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
		if rsid == takeoverCodexFixtureSessionID {
			foundMap = true
			if runner != "" && runner != "codex-tty" && runner != "codex" {
				t.Fatalf("auto-detect codex meta runner=%q want codex-tty/codex; meta=%s", runner, b)
			}
			break
		}
	}
	if !foundMap {
		t.Fatalf("no meta with runner_session_id=%s under %s; ids=%v",
			takeoverCodexFixtureSessionID, filepath.Join(req.Home, "sessions"), ids)
	}

	script := readItermScript(t, req)
	assertItermForceNewScript(t, script)
	assertContainsAny(t, script, "agent-run", "resume", "open", takeoverCodexFixtureSessionID)
	assertNoKillLog(t, req)
}
```
