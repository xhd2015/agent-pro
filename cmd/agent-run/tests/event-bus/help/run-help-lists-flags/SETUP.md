# Scenario

**Feature**: `agent-run run -h` lists `--event-bus-url` and `--event-bus-token`

```
agent-run run -h -> stdout contains both event-bus flags; trailing newline
```

## Steps

1. Set Args to `run -h`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.Args = []string{"run", "-h"}
	return nil
}
```
