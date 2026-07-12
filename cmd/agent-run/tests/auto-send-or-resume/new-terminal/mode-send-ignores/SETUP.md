# Scenario

**Feature**: live session + auto + `--new-terminal` ignores new-terminal and
still sends (MODE=send)

```
seed live sendable terminal
  -> run --auto-send-or-resume --new-terminal --session-id ID "followup"
  -> exit 0; stdout msg_N; no iTerm script (or empty)
  -> inject/enqueue works; optional stderr note about ignore
```

## Steps

1. Seed live bound not-exited session (fake ptywrap inject).
2. Enable iTerm script path so a mistaken OpenConfig would write a file.
3. Run auto + new-terminal + prompt.

```go
import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func Setup(t *testing.T, req *Request) error {
	req.SessionID = "nt-live-send-d3"
	req.RunnerSessionID = "550e8400-e29b-41d4-a716-446655440d33"
	req.TerminalSessionID = "term-nt-live-d3"
	req.InitialPrompt = "prior live nt"
	req.FollowupPrompt = "auto send despite new-terminal"
	seedLiveBoundNotExited(t, req)

	// If iTerm is wrongly invoked, script file will be non-empty.
	ensureItermScriptOutPath(req)
	_ = os.Remove(req.ItermScriptOut)

	// Argv probe reserved: send path must not spawn provider.
	req.ArgvProbePath = filepath.Join(req.TempDir, "argv-probe-should-not-exist.log")

	req.Args = []string{
		"run",
		"--auto-send-or-resume",
		"--new-terminal",
		"--session-id", req.SessionID,
		req.FollowupPrompt,
	}
	req.ExecTimeout = 45 * time.Second
	return nil
}
```
