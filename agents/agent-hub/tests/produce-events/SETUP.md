## Preconditions
- This group tests producing events via `agent-hub notify` and `agent-hub hook notify`.
- `AGENT_HUB_OPENCODE_RUNNER` may be set for runner redirection tests.

## Steps
1. Use `agent-hub notify` with `--json` or `--file` flags.
2. Use `agent-hub hook notify` with `--runner` and `--event` flags.
3. Use `fake-opencode run --mock-config` to simulate opencode sessions.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
    _ = t
    return nil
}
```
