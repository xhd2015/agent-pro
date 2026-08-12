# Scenario

**Feature**: resume footer in injected scrollback binds runner_session_id

```
unbound meta + SnapshotScrollback("… codex resume Y …")
  -> EnsureCodexRunnerBound -> runner_session_id=Y (return + store)
```

## Preconditions

- No matching rollouts required (scrollback is sufficient).
- Scrollback uses production footer phrase + `codex resume <uuid>`.

## Steps

1. Seed unbound codex-tty session.
2. Inject scrollback containing resume footer for id Y.
3. Call EnsureCodexRunnerBound; assert persist of Y.

```go
import (
	"fmt"
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.CreatedAt = "2026-08-11T10:00:00Z"
	req.RunnerSessionID = ""
	req.Rollouts = nil
	req.ScrollbackText = fmt.Sprintf(
		"CODEX_TTY_BANNER\nCodex › done\nTo continue this session, run codex resume %s\n",
		fixtureCodexIDOther,
	)
	return nil
}
```
