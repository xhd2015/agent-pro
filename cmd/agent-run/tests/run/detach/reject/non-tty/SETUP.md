# Scenario

**Feature**: `--detach` is rejected for non-TTY runners

```
agent-run run --detach --agent-runner fake-codex "x" -> exit ≠ 0
```

## Steps

1. Grouping marks non-TTY reject class.
2. Leaf picks a concrete non-TTY runner.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	if len(req.Args) == 0 || req.Args[0] != "run" {
		req.Args = []string{"run"}
	}
	return nil
}
```
