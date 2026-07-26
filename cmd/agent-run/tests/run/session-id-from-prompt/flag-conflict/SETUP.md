# Scenario

**Feature**: `--session` and `--session-id-from-prompt` are mutually exclusive

```
agent-run run --session X --session-id-from-prompt "prompt" → error, exit ≠ 0
```

## Preconditions

- Both flags present on the same `run` invocation.

## Steps

1. Grouping documents mutual-exclusion mode.
2. Leaf sets both flags with a non-empty prompt.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Runner = "fake-codex"
	// Prefix shared by conflict leaves; leaf adds session/auto/prompt.
	req.Args = []string{"run", "--agent-runner", "fake-codex"}
	return nil
}
```
