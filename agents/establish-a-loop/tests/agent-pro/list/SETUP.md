# Scenario

**Feature**: agent-pro skills lists establish-a-loop with description

```
agent-pro skills -> Available skills listing includes establish-a-loop + description
```

## Steps

1. Invoke `agent-pro skills` with no arguments.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Args = []string{"skills"}
	return nil
}
```