# Scenario

**Feature**: `agent-run run --help` lists `--session-id-from-prompt`

```
agent-run run --help → stdout contains --session-id-from-prompt; ends with newline
```

## Steps

1. Run `agent-run run --help` (args set by grouping).

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	// Args already: run --help from grouping.
	if len(req.Args) < 2 || req.Args[0] != "run" || req.Args[1] != "--help" {
		req.Args = []string{"run", "--help"}
	}
	return nil
}
```
