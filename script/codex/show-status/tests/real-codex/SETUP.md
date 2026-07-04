# Scenario

**Profile**: `real-codex` — production `codex` interactive TUI on PATH (`label: real-codex, slow`)

```
# no CODEX_SHOW_STATUS_COMMAND -> real codex in tty-watch -> /status -> live status lines
codex-show-status -> real codex -> status patterns (not exact values)
```

## Preconditions

- Real `codex` CLI must be on `PATH`; otherwise tests skip with `codex not found in PATH`.
- Do **not** set `CODEX_SHOW_STATUS_COMMAND` — production codex argv only.
- Leaves require `doctest test --label real-codex` or `--label slow` (excluded from default runs).
- Longer timeouts acceptable for real TUI startup.

## Steps

1. Grouping `Setup` calls `exec.LookPath("codex")` and sets `req.SkipFakeCommand = true`.
2. Leaf runs against live codex and asserts stdout line patterns.
3. Assert stderr is empty on success.

## Context

- Real-codex leaf asserts structure (`Monthly usage: <digits>%`, `Credits used: <digits> of <digits>`,
  `Next reset: <non-empty>`), not exact credit counts or reset date (live values change).

```go
import (
	"os/exec"
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	if _, err := exec.LookPath("codex"); err != nil {
		t.Skip("codex not found in PATH")
	}
	req.SkipFakeCommand = true
	req.TimeoutSeconds = "60"
	return nil
}
```