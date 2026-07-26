# Scenario

**Feature**: intent-route SKILL.md routes Investigation to investigate skill

```
agent-pro skill intent-route --show -> Investigation category + investigate guideline
```

## Steps

1. Invoke `agent-pro skill intent-route --show`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Args = []string{"skill", "intent-route", "--show"}
	return nil
}
```