# Scenario

**Subcommand**: `agent-run attach <session-id>` — registry lookup + interactive WS

```
agent-run attach <id>
  -> read grok-tty-registry/<id>.json
  -> ptyclient attach to hidden listen_addr
```

## Preconditions

- Attach only works while parent `run` is still blocking (registry file present).
- Unknown or expired ids exit 1 with actionable stderr.

## Steps

1. Grouping `Setup` documents attach subcommand context.
2. Leaf `Setup` either runs attach against missing id or starts background run first.
3. `Run` executes attach CLI or WS probe via registry.
4. `Assert` checks connection success or missing-session error.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = t
	return nil
}
```