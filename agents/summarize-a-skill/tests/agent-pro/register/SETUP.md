# Scenario

**Feature**: knownSkills includes summarize-a-skill

```
agent-pro skill summarize-a-skill --show -> name: summarize-a-skill in output
```

## Steps

1. Invoke `agent-pro skill summarize-a-skill --show`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Args = []string{"skill", "summarize-a-skill", "--show"}
	return nil
}
```
