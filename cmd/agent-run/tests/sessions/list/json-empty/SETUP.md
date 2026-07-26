# Scenario

**Feature**: `sessions --json` on empty store returns valid empty JSON

```
agent-run sessions --json (empty AGENT_RUN_HOME) → {"sessions":[]}
```

## Preconditions

- Fresh `AGENT_RUN_HOME` has no sessions.

## Steps

1. Run `agent-run sessions --json` on an empty `AGENT_RUN_HOME`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Args = append(req.Args, "--json")
	return nil
}
```
