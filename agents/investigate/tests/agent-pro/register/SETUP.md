# Scenario

**Feature**: knownSkills includes investigate

```
agent-pro skill investigate --show -> name: investigate in output
```

## Steps

1. Invoke `agent-pro skill investigate --show`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Args = []string{"skill", "investigate", "--show"}
	return nil
}
```