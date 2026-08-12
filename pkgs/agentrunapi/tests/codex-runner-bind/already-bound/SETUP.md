# Scenario

**Feature**: already-bound runner_session_id is never overwritten

```
meta.runner_session_id=A + scrollback/discovery offer B
  -> EnsureCodexRunnerBound no-op; stay A; bound=true
```

## Preconditions

- Meta starts with non-empty RunnerSessionID (A).
- Scrollback and a matching rollout both offer a different id (B).

## Steps

1. Seed bound session with id A.
2. Offer id B via scrollback and matching rollout.
3. Assert return + store still A (no overwrite).

```go
import (
	"fmt"
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.CreatedAt = "2026-08-11T10:00:00Z"
	req.RunnerSessionID = fixtureCodexIDBound
	req.ScrollbackText = fmt.Sprintf(
		"To continue this session, run codex resume %s\n",
		fixtureCodexIDOffer,
	)
	req.Rollouts = []RolloutSeed{
		{
			CodexSessionID: fixtureCodexIDOffer,
			Cwd:            req.Workspace,
			Timestamp:      "2026-08-11T12:32:25Z",
		},
	}
	return nil
}
```
