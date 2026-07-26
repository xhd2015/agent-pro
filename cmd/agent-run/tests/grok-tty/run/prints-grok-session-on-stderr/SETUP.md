# Scenario

**Bug**: stderr omits discovered grok session id and updates.jsonl path after internal session id

```
temp GROK_HOME session with known UUID
  -> stderr grok-tty: session-N
  -> stderr grok-tty: grok session <uuid>
  -> stderr grok-tty: grok updates <absolute-path>
```

## Steps

1. Seed fake grok session dir with fixed UUID and matching prompt.
2. Run with short-hold fake TUI so discovery completes during PTY lifetime.
3. Assert stderr contains grok session id and updates path after discovery.

```go
import (
	"path/filepath"
	"testing"

	"github.com/xhd2015/doctest/session"
)

const stderrGrokUUID = "550e8400-e29b-41d4-a716-446655440000"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.GrokHome = filepath.Join(req.TempDir, "grok-home")
	req.GrokSessionUUID = stderrGrokUUID
	prompt := "stderr grok session"
	req.GrokUpdatesPath = writeFakeGrokSessionDir(t, req.GrokHome, req.TempDir, stderrGrokUUID, prompt,
		acpAgentMessageChunk("STDERR_GROK_ASSISTANT"),
	)
	appendGrokHomeEnv(req)

	req.GrokTTYCommand = fakeTUIHoldSeconds(2)
	appendGrokTTYEnv(req)
	req.Args = append(req.Args, prompt)
	return nil
}
```