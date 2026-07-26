# Scenario

**Feature**: `agent-run resume --help` lists `--detach`

```
agent-run resume --help → stdout contains --detach; ends with newline
```

## Steps

1. Run `agent-run resume --help`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Args = []string{"resume", "--help"}
	return nil
}
```
