# Scenario

**Feature**: live codex returns parseable status line patterns

```
# real codex TUI -> /status -> Monthly usage + Credits used + Next reset
codex-show-status -> real codex -> pattern assertions
```

## Preconditions

- Real `codex` on PATH (inherited from `real-codex/SETUP.md`).
- No fake TUI hook.

## Steps

1. Run CLI against production codex argv.
2. Assert exit 0; stdout matches status line patterns; stderr empty.

## Context

- Optional leaf for manual/CI-with-codex verification; skipped without `--label real-codex`.

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