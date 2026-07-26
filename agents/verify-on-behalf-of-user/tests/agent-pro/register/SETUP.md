# Scenario

**Feature**: knownSkills includes verify-on-behalf-of-user with nested topics

```
agent-pro skill verify-on-behalf-of-user transcript --show -> transcript topic
```

## Steps

1. Invoke `agent-pro skill verify-on-behalf-of-user transcript --show`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Args = []string{"skill", "verify-on-behalf-of-user", "transcript", "--show"}
	return nil
}
```