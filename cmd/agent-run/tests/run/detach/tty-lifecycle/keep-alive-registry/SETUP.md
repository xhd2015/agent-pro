# Scenario

**Feature**: `--detach` implies keep-alive — registry entry remains reachable after parent exits

```
agent-run run --agent-runner grok-tty --detach "hi"
  -> exit 0
  -> registry/<terminal-id>.json exists
  -> listen_addr TCP still open (reattach/send possible)
```

## Preconditions

- Fake TUI holds long enough that the keep-alive server outlives the parent CLI.

## Steps

1. Run detach lifecycle to completion.
2. Mode `detach-registry-after` loads registry for the printed terminal id.
3. Assert registry file + TCP reachability.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Prompt = "hi"
	req.Mode = "detach-registry-after"
	req.Args = []string{"run", "--agent-runner", "grok-tty", "--detach", req.Prompt}
	setGrokTTYCommand(req, fakeTUIHoldSeconds(45))
	return nil
}
```
