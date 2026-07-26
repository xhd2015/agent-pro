# Scenario

**Feature**: `--detach` with `fake-codex` fails as non-TTY

```
agent-run run --agent-runner fake-codex --detach "x"
  -> exit ≠ 0
  -> stderr explains --detach requires a TTY runner
```

## Steps

1. Invoke run with `--detach`, `fake-codex`, and a short prompt.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Runner = "fake-codex"
	req.Prompt = "x"
	req.Args = []string{"run", "--agent-runner", "fake-codex", "--detach", req.Prompt}
	return nil
}
```
