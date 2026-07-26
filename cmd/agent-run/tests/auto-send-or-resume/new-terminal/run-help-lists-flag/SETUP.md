# Scenario

**Feature**: `agent-run run -h` documents `--new-terminal`

```
agent-run run -h -> stdout contains --new-terminal; trailing newline
```

## Steps

1. Run `agent-run run -h`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.ItermScriptOut = ""
	req.Args = []string{"run", "-h"}
	return nil
}
```
