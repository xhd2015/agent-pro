# Scenario

**Feature**: `AGENT_RUN_HOME` overrides explicit constructor home

```
NewFileStore(ignoredPath) + AGENT_RUN_HOME=override -> Store.Home() == override
```

## Preconditions

- Env var `AGENT_RUN_HOME` is set to a path different from `req.Home`.
- Store must resolve home from the environment, not the constructor argument.

## Steps

1. Set `req.Action = "env_override"`.
2. Point `AGENT_RUN_HOME` at `filepath.Join(req.TempDir, "override-home")`.
3. Keep `req.Home` at the default `.agent-run` path (constructor arg).
4. Open store and read `Store.Home()`.

```go
import (
	"path/filepath"
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	override := filepath.Join(req.TempDir, "override-home")
	req.Env = []string{"AGENT_RUN_HOME=" + override}
	req.Action = "env_override"
	return nil
}
```