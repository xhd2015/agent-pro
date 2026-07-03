# Scenario

**Feature**: default fake grok returns fixture usage lines

```
# default fake TUI -> /usage show -> Weekly limit: 1% + Next reset: July 9, 16:55 PT
grok-show-usage -> fake grok -> usage lines
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
	req.ShowUsageCommand = fakeTUIDefault()
	return nil
}
```