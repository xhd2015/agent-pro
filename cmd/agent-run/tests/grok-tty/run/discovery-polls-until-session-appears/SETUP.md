# Scenario

**Bug**: discovery must keep polling until delayed grok session dir appears (no 30s hard cap)

```
no grok session dir at start
  -> prompt injected after banner
  -> session dir + updates.jsonl created 5s later
  -> discovery still finds session and streams DELAYED_SESSION_MARKER
```

## Steps

1. Point `GROK_HOME` at temp dir; do **not** set `AGENT_RUN_GROK_TTY_GROK_SESSION_ID`.
2. Schedule delayed session dir creation 5s after internal session id appears.
3. Fake TUI holds 20s so discovery can outlive the old 30s cap window.
4. `Mode=stream-probe` waits for `DELAYED_SESSION_MARKER`.

```go
import (
	"testing"
	"time"
)

const (
	delayedSessionGrokUUID = "cccccccc-cccc-cccc-cccc-cccccccccccc"
	delayedSessionMarker   = "DELAYED_SESSION_MARKER"
	delayedSessionPrompt   = "delayed discovery probe"
)

func Setup(t *testing.T, req *Request) error {
	req.GrokHome = filepath.Join(req.TempDir, "grok-home")
	appendGrokHomeEnv(req)
	req.Env = withoutEnvKey(req.Env, "AGENT_RUN_GROK_TTY_GROK_SESSION_ID")

	sched, updatesPath := delayedGrokSessionSchedule(t, 5*time.Second, req.GrokHome, req.TempDir, delayedSessionGrokUUID, delayedSessionPrompt,
		acpAgentMessageChunk(delayedSessionMarker),
	)
	req.GrokUpdatesPath = updatesPath
	req.GrokUpdatesSchedules = []GrokUpdatesSchedule{sched}

	req.GrokTTYCommand = fakeTUIHoldSeconds(20)
	appendGrokTTYEnv(req)
	req.Args = append(req.Args, delayedSessionPrompt)

	req.Mode = "stream-probe"
	req.StreamProbeSubstring = delayedSessionMarker
	req.StreamProbeTimeout = 18 * time.Second
	return nil
}
```