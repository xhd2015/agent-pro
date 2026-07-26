# Scenario

**Feature**: `agent-run run -h` documents `--auto-send-or-resume`

```
agent-run run -h -> stdout contains --auto-send-or-resume; trailing newline
```

## Steps

1. Run `agent-run run -h`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Args = []string{"run", "-h"}
	return nil
}
```
