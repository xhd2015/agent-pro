# Scenario

**Feature**: `--open` with `fake-codex` fails as non-TTY

```
agent-run run --agent-runner fake-codex --open "x"
  -> exit ≠ 0
  -> stderr explains --open requires a TTY runner (not unrecognized flag only)
```

## Steps

1. Invoke run with `--open`, `fake-codex`, and a short prompt.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Runner = "fake-codex"
	req.Prompt = "x"
	req.Args = []string{"run", "--agent-runner", "fake-codex", "--open", req.Prompt}
	return nil
}
```
