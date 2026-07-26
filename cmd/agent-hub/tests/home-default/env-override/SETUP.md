# Scenario

**Feature**: AGENT_HUB_HOME on child env overrides default home

```
# child Env: AGENT_HUB_HOME=<custom> + HOME
agent-hub daemon status -> JSON home = custom path
```

## Preconditions

- `AGENT_HUB_HOME` is set on the child to a custom temp directory.
- Env takes precedence over the default under `HOME`.

## Steps

1. Append `AGENT_HUB_HOME` to `req.Env` (child only).
2. Run `agent-hub daemon status`.
3. Confirm `home` matches the custom path.

```go
import (
	"path/filepath"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	customHome := filepath.Join(req.TempDir, "custom-agent-hub")
	req.Env = append(req.Env, "AGENT_HUB_HOME="+customHome)
	req.Args = []string{"daemon", "status"}
	return nil
}
```
