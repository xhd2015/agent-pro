# Scenario

**Feature**: default fake codex returns fixture status lines

```
# default fake TUI -> /status -> Monthly usage 58% + credits + reset
codex-show-status -> fake codex -> status lines
```

## Preconditions

- Default `fakeTUIDefault()` hook from success grouping.

## Steps

1. Inherit default fake TUI from `success/SETUP.md`.
2. Run and assert exact stdout fixture.

## Context

- Baseline happy-path test for the CLI.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.ShowStatusCommand = fakeTUIDefault()
	return nil
}
```