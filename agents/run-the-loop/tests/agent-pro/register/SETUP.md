# Scenario

**Feature**: knownSkills includes run-the-loop

```
agent-pro skill run-the-loop --show -> name: run-the-loop in output
```

## Steps

1. Invoke `agent-pro skill run-the-loop --show`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Args = []string{"skill", "run-the-loop", "--show"}
	return nil
}
```