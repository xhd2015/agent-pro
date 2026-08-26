# Scenario

**Feature**: --set-config --agent-runner persists to config.json

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Args = []string{"--set-config", "--agent-runner", "codex"}
	return nil
}
```
