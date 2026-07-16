# Scenario

**Feature**: `status --grok-session-id ID` resolves meta by `runner_session_id`
when `meta.runner` is exactly `grok` or `grok-tty`

```
seed finished bound meta (runner=grok|grok-tty, runner_session_id=UUID)
  -> agent-run status --grok-session-id UUID
  -> exit 0; multi-layer view for that agent-run session
  # 0 matches → not found; 2+ → ambiguous; codex* never match
  # --grok-session-id and positional are mutually exclusive
```

## Steps

1. Leaf seeds meta (and optional second session for ambiguity).
2. Leaf sets `req.Args` to `status --grok-session-id …` (and error variants).

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	// Grouping for --grok-session-id lookup leaves under status.
	if req.Runner == "" {
		req.Runner = defaultRunner
	}
	return nil
}
```
