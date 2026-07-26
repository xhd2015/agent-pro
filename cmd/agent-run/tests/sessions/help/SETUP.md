# Scenario

**Feature**: `sessions --help` documents list/print options including
`--grok-session-id`

```
agent-run sessions --help -> options + --grok-session-id
```

## Steps

1. Leaf runs `sessions --help` and asserts documented flags.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	if len(req.Args) == 0 {
		req.Args = []string{"sessions", "--help"}
	}
	return nil
}
```
