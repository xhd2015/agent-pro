# Scenario

**Feature**: knownSkills includes debug-with-user

```
agent-pro skill debug-with-user --show -> name: debug-with-user in output
```

## Steps

1. Invoke `agent-pro skill debug-with-user --show`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Args = []string{"skill", "debug-with-user", "--show"}
	return nil
}
```
