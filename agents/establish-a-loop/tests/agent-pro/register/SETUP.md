# Scenario

**Feature**: knownSkills includes establish-a-loop

```
agent-pro skill establish-a-loop --show -> name: establish-a-loop in output
```

## Steps

1. Invoke `agent-pro skill establish-a-loop --show`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Args = []string{"skill", "establish-a-loop", "--show"}
	return nil
}
```