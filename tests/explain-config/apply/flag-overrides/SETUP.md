# Scenario

**Feature**: CLI --agent-runner overrides persisted config

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
	writeConfigJSON(t, req.ConfigHome, "{\n  \"version\": 1,\n  \"agent_runner\": \"codex\"\n}\n")
	req.Args = []string{"--agent-runner", "opencode", "hello override"}
	return nil
}
```
