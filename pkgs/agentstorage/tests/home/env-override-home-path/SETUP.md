# Scenario

**Feature**: `AGENT_RUN_HOME` applies when constructor home is empty

```
NewFileStore("") + AGENT_RUN_HOME=override -> Store.Home() == override
```

## Preconditions

- Env var `AGENT_RUN_HOME` is set to a path different from `req.Home`.
- Constructor home is empty so product resolveHome uses AGENT_RUN_HOME.
- Residual process Setenv only for this env-contract leaf.

## Steps

1. Set `req.Action = "env_override"`.
2. Point `AGENT_RUN_HOME` at `filepath.Join(req.TempDir, "override-home")`.
3. Harness opens with empty constructor home (see openStore).
4. Open store and read `Store.Home()`.

```go
import (
	"path/filepath"
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	override := filepath.Join(req.TempDir, "override-home")
	req.Env = []string{"AGENT_RUN_HOME=" + override}
	req.Action = "env_override"
	return nil
}
```