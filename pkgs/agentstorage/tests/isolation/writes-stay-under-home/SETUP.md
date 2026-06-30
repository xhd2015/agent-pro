# Scenario

**Feature**: all store writes remain under resolved home directory

```
CreateSession + SaveConfig + AppendEvent + AppendMessage -> scan tree -> AssertHomeOnly
```

## Preconditions

- Fresh isolated `AGENT_RUN_HOME`.
- Multiple write paths exercised in one run.

## Steps

1. Inherited from grouping `isolation/SETUP.md` (`writes_under_home` action).
2. No additional leaf-specific configuration required.

```go
import (
	"testing"

	"github.com/xhd2015/agent-pro/pkgs/agentstorage"
)

func Setup(t *testing.T, req *Request) error {
	req.Operation = "isolation"
	req.Action = "writes_under_home"
	req.SessionID = "sess_iso_leaf"
	req.Config = agentstorage.Config{
		DefaultAgentRunner: "fake-opencode",
		DefaultModel:       "isolation-model",
		LastSession:        "fake-opencode/sess_iso_leaf",
	}
	return nil
}
```