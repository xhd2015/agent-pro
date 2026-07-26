# Scenario

**Feature**: knownSkills includes verify-with-prototype

```
agent-pro skill verify-with-prototype --show -> name: verify-with-prototype in output
```

## Steps

1. Invoke `agent-pro skill verify-with-prototype --show`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Args = []string{"skill", "verify-with-prototype", "--show"}
	return nil
}
```