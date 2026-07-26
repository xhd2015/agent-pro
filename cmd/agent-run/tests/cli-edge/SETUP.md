# Scenario

**Subcommand**: `cli-edge` — invalid subcommands and unknown agent runners

## Preconditions

- Short-path leaves use L2 `req.Mode = "handle"` (`agentruncli.Handle`); no binary.
- Tests expect non-zero exit codes and actionable stderr messages.

## Steps

1. Leaf `Setup` sets `Mode: "handle"` and `req.Args` for the invalid invocation.
2. `Run` calls Handle in-process (maps error → exit 1 / stderr like thin main).
3. `Assert` checks exit code 1 and stderr content.

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
