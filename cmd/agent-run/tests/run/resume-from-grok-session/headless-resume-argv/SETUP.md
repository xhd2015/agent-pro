# Scenario

**Feature**: valid `--resume-from-grok-session` headless launch pre-binds the
Grok UUID and spawns the provider with `--resume <uuid>` (argv-recorder)

```
seed Grok UUID under GROK_HOME; no agent-run mapping
argv-recorder as --agent-runner-binary
  -> agent-run run --session-id FIXED --agent-runner-binary REC
       --resume-from-grok-session UUID "followup"
  -> exit 0
  -> ARGV_RECORD contains --resume UUID
```

## Preconditions

- P1 validation passes (id present, runner ok, not mapped, dir ok).
- `AGENT_RUN_GROK_TTY_COMMAND` unset so recorder sees real argv.
- Fixed `--session-id` for deterministic store path (optional but preferred).

## Steps

1. Seed Grok session at process workspace.
2. Install argv-recording fake runner.
3. Run import with fixed session id, binary, and non-empty followup.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.SessionID = "import-headless-argv-1"
	req.FollowupPrompt = "resume-from-grok followup"
	setupValidImport(t, req, true)
	req.Args = runArgs(req, req.GrokSessionID, req.FollowupPrompt)
	return nil
}
```
