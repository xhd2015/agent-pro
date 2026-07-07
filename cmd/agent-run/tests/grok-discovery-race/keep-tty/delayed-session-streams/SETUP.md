# Scenario

**Bug**: delayed grok session must stream to events.jsonl under keep-tty + chrome (no early cancel)

```
empty GROK_HOME at start + real-like chrome + --keep-tty
  -> updates.jsonl materializes 8s after grok-tty session id
  -> discovery keeps polling until streamed marker appears
```

Mirrors `script/debug/grok-tty-discovery-cancel -scenario=delayed-session` and
`grok-tty/run/discovery-polls-until-session-appears` with keep-tty + llm-mock chrome.

## Steps

1. Point `GROK_HOME` at empty temp dir; do not pre-seed session dirs.
2. Schedule delayed session dir + `DELAYED_SESSION_MARKER` 8s after internal session id.
3. Configure `LLM_MOCK_RUN_GROK_COMMAND` real-like chrome holding 30s.
4. Run with fixed `--session` id; poll `events.jsonl` for marker.

```go
import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func Setup(t *testing.T, req *Request) error {
	req.Scenario = "delayed-session-streams"
	req.SessionID = "grok_discovery_delayed"
	req.Prompt = chromeWaitPrompt
	req.GrokHome = filepath.Join(req.TempDir, "grok-home")
	if err := os.MkdirAll(req.GrokHome, 0755); err != nil {
		return err
	}
	stripEnvPrefix(req, "GROK_HOME=")
	req.Env = append(req.Env, "GROK_HOME="+req.GrokHome)

	sched, updatesPath := delayedGrokSessionSchedule(t, 8*time.Second, req.GrokHome, req.TempDir, delayedSessionGrokUUID, req.Prompt,
		acpAgentMessageChunk(delayedSessionMarker),
	)
	req.GrokUpdatesPath = updatesPath
	req.GrokUpdatesSchedules = []GrokUpdatesSchedule{sched}

	configureLLMMockChromeEnv(t, req, req.Prompt, req.ChromeHoldSeconds)
	req.ExecTimeout = 45 * time.Second
	return nil
}
```