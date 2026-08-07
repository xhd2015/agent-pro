# Scenario

**Feature**: `run --color` forces TTY child color env (policy last)

```
agent-run run --agent-runner grok-tty --color \
  --agent-runner-binary <env-logger> "prompt"
  -> child env force (NO_COLOR unset; FORCE/CLICOLOR trio; TERM fixup)
```

## Steps

1. Grouping prefixes common TTY run with `--color`.
2. Leaves set parent env factors, env-logger, session id, finalize args.
3. Assert probe (or non-TTY hard error).

```go
import (
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Color = true
	// Default TTY run prefix with --color; leaves append binary/prompt/session.
	req.Args = []string{"run", "--agent-runner", "grok-tty", "--color"}
	return nil
}
```
