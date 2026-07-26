# Scenario

**Feature**: CLI help documents `--detach`

```
agent-run run --help -> --detach
agent-run resume --help -> --detach
```

## Preconditions

- `agent-run` binary is built (root Setup).

## Steps

1. Grouping marks help mode (no runner required).
2. Leaves run `run --help` or `resume --help`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Runner = ""
	return nil
}
```
