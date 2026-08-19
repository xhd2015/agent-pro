# Scenario

**Bug**: post-turn `agent-run snapshot` omits grok TUI conversation screen

```
fake grok TUI simulating llm-mock post-turn redraw (status bar only)
  -> background grok-tty run --keep-tty
  -> wait for prompt stream marker
  -> agent-run snapshot must show prompt + grok UI (tty-watch parity)
```

Phase A with real `llm-mock-run-grok` showed snapshot output:

```
GROK_HOME=...
     Turn completed in 5.3s.
                     Ctrl+.:shortcuts
```

while `tty-watch snapshot` on an equivalent session includes `one word of France captial`
and grok home/menu lines. This leaf replays that post-turn PTY scrollback with a
deterministic fake TUI.

## Steps

1. Start background `agent-run run --agent-runner grok-tty --keep-tty` with
   `fakeGrokPostTurnSnapshotTUI()`.
2. Wait for streamed stdout marker, sleep 2s, run `agent-run snapshot <session-id>`.

```go
import (
	"path/filepath"
	"testing"

	"github.com/xhd2015/doctest/session"
)

const snapshotMockPrompt = "one word of France captial"

func fakeGrokPostTurnSnapshotTUI() string {
	return `sh -c 'printf "GROK_TTY_BANNER\nGrok Build Beta\n"; sleep 0.5; printf "one word of France captial\n               New worktree                                 ctrl+w\n               Resume session                               ctrl+s\n               Changelog\n               Quit                                              q\n"; sleep 30; printf "\x1b[2J\x1b[H     Turn completed in 5.3s.                                                   █\n                                                                               █\n                     Ctrl+.:shortcuts\n"; sleep 120'`
}

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.GrokHome = filepath.Join(req.TempDir, "grok-home")
	req.GrokSessionUUID = "eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee"
	appendGrokHomeEnv(req)
	_ = writeFakeGrokSessionDir(t, req.GrokHome, req.TempDir, req.GrokSessionUUID, snapshotMockPrompt,
		acpAgentMessageChunk("Done! I've addressed your request."),
	)
	req.GrokTTYCommand = fakeGrokPostTurnSnapshotTUI()
	appendGrokTTYEnv(req)
	req.KeepTTY = true
	req.GrokTTYPrompt = snapshotMockPrompt
	startGrokTTYBackground(t, req)
	req.Mode = "snapshot-probe"
	req.SnapshotReadyMarker = "Done!"
	return nil
}
```