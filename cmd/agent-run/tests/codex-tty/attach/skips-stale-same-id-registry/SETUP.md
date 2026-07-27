# Scenario

**Bug**: `agent-run attach` skips a dead grok-tty registry entry that shares the live codex-tty id

```
stale grok-tty-registry/session-1.json -> refused localhost port
live codex-tty-registry/session-1.json -> running ptywrap listener

agent-run attach session-1 -> skip stale grok candidate -> attach live codex candidate
```

## Preconditions

- The stale grok-tty entry is written before the live codex-tty background run.
- The codex-tty run uses an isolated `AGENT_RUN_HOME`, so its first PTY session id
  should be `session-1`.

## Steps

1. Write `grok-tty-registry/session-1.json` pointing to an unused localhost port.
2. Start a long-running fake codex-tty session in the background.
3. Verify the live codex-tty session id is also `session-1`.
4. Run `agent-run attach session-1` through the CLI, without direct registry fallback.

```go
import (
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	writeStaleGrokTTYRegistry(t, req.Home, "session-1")
	req.CodexTTYCommand = fakeTUILongRun()
	req.CodexTTYPrompt = "hold"
	startCodexTTYBackground(t, req)
	if req.CodexTTYSessionID != "session-1" {
		t.Fatalf("same-id collision requires live codex session-1, got %q", req.CodexTTYSessionID)
	}
	req.Mode = "attach-cli-only-probe"
	return nil
}
```
