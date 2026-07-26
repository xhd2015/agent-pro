# Scenario

**Feature**: `sessions --help` documents `--grok-session-id`

```
agent-run sessions --help -> mentions --grok-session-id (and --print)
```

## Steps

1. Run `agent-run sessions --help`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Args = []string{"sessions", "--help"}
	return nil
}
```
