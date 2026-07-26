# Scenario

**Feature**: `run --help` documents `--session-id-from-prompt`

```
agent-run run --help -> usage lists --session-id-from-prompt
```

## Preconditions

- `agent-run` binary is built (root Setup).

## Steps

1. Grouping marks help mode (no runner required).
2. Leaf runs `run --help` and asserts flag text.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	// Help path: no agent-runner; clear any accidental runner defaults.
	req.Runner = ""
	req.Args = []string{"run", "--help"}
	return nil
}
```
