# Scenario

**Feature**: with AGENT_HUB_HOME unset, daemon status home is $HOME/.agent-hub

```
# child Env has HOME only (no AGENT_HUB_HOME)
agent-hub daemon status -> JSON home = $HOME/.agent-hub, running=false
```

## Preconditions

- `AGENT_HUB_HOME` is unset on the child.
- `HOME` points to the isolated temp user home from root Setup.

## Steps

1. Run `agent-hub daemon status` without `AGENT_HUB_HOME`.
2. The output JSON `home` field reveals the default path.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Args = []string{"daemon", "status"}
	return nil
}
```
