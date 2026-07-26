# Scenario

**Feature**: `agent-run run --help` lists `--detach`

```
agent-run run --help → stdout contains --detach; ends with newline
```

## Steps

1. Run `agent-run run --help`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Args = []string{"run", "--help"}
	return nil
}
```
