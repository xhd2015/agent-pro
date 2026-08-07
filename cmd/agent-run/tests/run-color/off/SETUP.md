# Scenario

**Feature**: without `--color`, agent-run does not force color env / TERM rewrite

```
run --agent-runner grok-tty --agent-runner-binary env-logger "prompt"
  (no --color; hostile parent NO_COLOR / TERM=dumb / FORCE_COLOR=0)
  -> child keeps baseline; no FORCE_COLOR=1 force from this feature
  -> TERM not rewritten from dumb
```

## Steps

1. Grouping prefixes TTY run **without** `--color`.
2. Leaf sets hostile parent env + env-logger.

```go
import (
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Color = false
	req.Args = []string{"run", "--agent-runner", "grok-tty"}
	return nil
}
```
