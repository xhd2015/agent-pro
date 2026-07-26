# Scenario

**Feature**: knownSkills includes consolidate-code

```
agent-pro skill consolidate-code --show -> name: consolidate-code in output
```

## Steps

1. Invoke `agent-pro skill consolidate-code --show`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Args = []string{"skill", "consolidate-code", "--show"}
	return nil
}
```