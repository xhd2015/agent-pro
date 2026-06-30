# Scenario

**Feature**: config save and reload preserves all fields

```
SaveConfig(cfg) -> Config() -> cfg fields equal
```

## Preconditions

- Home is fresh; first write creates `config.json`.

## Steps

1. Set `req.Action = "save_reload"`.
2. Populate `req.Config` with non-empty runner, model, and last session.

```go
import (
	"testing"

	"github.com/xhd2015/agent-pro/pkgs/agentstorage"
)

func Setup(t *testing.T, req *Request) error {
	req.Action = "save_reload"
	req.Config = agentstorage.Config{
		DefaultAgentRunner: "fake-codex",
		DefaultModel:       "gpt-test",
		LastSession:        "fake-codex/sess_last",
	}
	return nil
}
```