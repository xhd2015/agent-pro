# Scenario

**Feature**: TTY auto-id collision against existing storage appends `-N` and keeps same-id policy

```
seed sessions/grok-tty/hello-world-<nearby-ts>/
agent-run run --agent-runner grok-tty --keep-tty --session-id-from-prompt "hello world"
  -> stderr + storage + registry share hello-world-<ts>-N
```

## Steps

1. Seed storage collisions for base `hello-world` under `grok-tty`.
2. Run auto-id with prompt `hello world`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Prompt = "hello world"
	seedStorageCollisionsForBase(t, req.Home, "grok-tty", "hello-world")
	req.Args = append(req.Args, req.Prompt)
	return nil
}
```
