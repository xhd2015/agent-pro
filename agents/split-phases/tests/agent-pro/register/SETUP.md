# Scenario

**Feature**: knownSkills includes split-phases

```
agent-pro skill split-phases --show -> name: split-phases in output
```

## Steps

1. Invoke `agent-pro skill split-phases --show`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Args = []string{"skill", "split-phases", "--show"}
	return nil
}
```
