# Scenario

**Feature**: `status --help` documents session-ref, multi-layer probe, and
`--grok-session-id`

```
agent-run status --help -> session id / layers; --grok-session-id
```

## Steps

1. Run `agent-run status --help`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Args = []string{"status", "--help"}
	return nil
}
```
