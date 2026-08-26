# Scenario

**Feature**: persisted agent_runner is used when CLI omits --agent-runner

Seeds config with codex, runs with -v and failing fake so Start fails after
verbose names the runner.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	writeConfigJSON(t, req.ConfigHome, "{\n  \"version\": 1,\n  \"agent_runner\": \"codex\"\n}\n")
	req.Args = []string{"-v", "hello from config"}
	return nil
}
```
