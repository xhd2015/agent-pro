# Scenario

**Feature**: `run` blocks until the PTY command exits

```
fake TUI exits after echo → agent-run run returns exit 0
```

## Steps

1. Run with fake TUI respond script; prompt `ping`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.GrokTTYCommand = fakeTUIRespondHi()
	req.Args = append(req.Args, "ping")
	return nil
}
```