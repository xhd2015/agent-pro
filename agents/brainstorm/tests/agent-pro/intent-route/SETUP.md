# Scenario

**Feature**: intent-route SKILL.md routes Flash Ideas to brainstorm skill

```
agent-pro skill intent-route --show -> Flash Idea category + brainstorm guideline
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