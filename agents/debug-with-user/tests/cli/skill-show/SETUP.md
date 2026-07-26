# Scenario

**Feature**: skill show prints embedded SKILL.md for agents

```
debug-with-user skill show -> stdout contains YAML frontmatter name
```

## Steps

1. Run `skill show` with no extra arguments.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Args = []string{"skill", "show"}
	return nil
}
```
