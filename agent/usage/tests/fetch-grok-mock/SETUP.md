# Scenario

**Feature**: usage.Fetch Grok provider with GROK_SHOW_USAGE_COMMAND fake TUI

```
usage.Fetch(ctx, Grok) -> grok tty -> fake TUI -> Snapshot{Provider:grok, UsagePercent:1%, Reset:...}
```

## Steps

1. Set `req.Provider` to `usage.Grok`.
2. Set `req.ShowUsageCommand` to default fake grok TUI fixture.

```go
import (
	"testing"

	"github.com/xhd2015/agent-pro/agent/usage"
)

func Setup(t *testing.T, req *Request) error {
	req.Provider = usage.Grok
	req.ShowUsageCommand = fakeGrokTUIDefault()
	return nil
}
```