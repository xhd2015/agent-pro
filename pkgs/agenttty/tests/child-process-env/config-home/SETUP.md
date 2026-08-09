# Scenario

**Feature**: configHome produces runner-specific home env assignment

```
# config home
configHome=/tmp/runner-home + runnerID
  -> CODEX_HOME or GROK_HOME in Set
```

## Steps

1. Grouping sets a fixed ConfigHome and Color=false.
2. Leaves choose RunnerID codex-tty vs grok-tty.

## Context

- Empty configHome is covered by empty-policy; this group requires non-empty home.
- Matches `RunnerConfigHomeEnv` mapping.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

const testConfigHome = "/tmp/agent-pro-child-env-config-home"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.Mode = "build"
	req.Color = false
	req.ConfigHome = testConfigHome
	req.PrependPaths = nil
	req.EnvEntries = nil
	req.ParentTERM = ""
	return nil
}
```
