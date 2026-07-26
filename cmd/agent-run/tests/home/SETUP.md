# Scenario

**Subcommand**: `home` — `AGENT_RUN_HOME` isolation and storage layout

## Preconditions

- `agent-run` binary is built (inherited from root `SETUP.md`).
- All durable writes must stay under `AGENT_RUN_HOME`.

## Steps

1. Leaf `Setup` runs a headless `run` that creates session files.
2. `Run` executes `agent-run` (inherited from leaf).
3. `Assert` walks `req.TempDir` and verifies no files exist outside `req.Home`.

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