# Scenario

**Feature**: non-codex runners never bind via EnsureCodexRunnerBound

```
runner=grok-tty + matching rollout + resume footer
  -> EnsureCodexRunnerBound no-op; stay unbound
```

## Preconditions

- Runner is `grok-tty` (not codex / codex-tty / codex-*).
- Both discovery and scrollback would bind if runner were codex.

## Steps

1. Seed unbound grok-tty session with workspace W.
2. Offer matching rollout + resume footer with Codex id X.
3. Assert stay unbound (no false bind / no panic).

```go
import (
	"fmt"
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Runner = "grok-tty"
	req.CreatedAt = "2026-08-11T10:00:00Z"
	req.RunnerSessionID = ""
	req.ScrollbackText = fmt.Sprintf(
		"To continue this session, run codex resume %s\n",
		fixtureCodexIDMatching,
	)
	req.Rollouts = []RolloutSeed{
		{
			CodexSessionID: fixtureCodexIDMatching,
			Cwd:            req.Workspace,
			Timestamp:      "2026-08-11T12:32:25Z",
		},
	}
	return nil
}
```