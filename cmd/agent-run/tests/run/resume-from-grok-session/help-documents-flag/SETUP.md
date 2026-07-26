# Scenario

**Feature**: `agent-run run --help` documents `--resume-from-grok-session`

```
agent-run run --help
  -> exit 0
  -> stdout contains --resume-from-grok-session
  -> stdout ends with trailing newline
```

## Steps

1. Run `agent-run run --help` (no GROK fixture required).

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
