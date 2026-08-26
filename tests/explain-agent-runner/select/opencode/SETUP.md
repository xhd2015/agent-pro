# Scenario

**Feature**: default opencode runner persists session with working fake agent

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	working, err := ensureWorkingFakeOpencode(t, d)
	if err != nil {
		return err
	}
	req.WorkingAgentPath = working
	req.Args = []string{"hello from doctest"}
	return nil
}
```
