# Scenario

**Feature**: knownSkills includes brainstorm

```
agent-pro skill brainstorm --show -> name: brainstorm in output
```

## Steps

1. Invoke `agent-pro skill brainstorm --show`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Args = []string{"skill", "brainstorm", "--show"}
	return nil
}
```