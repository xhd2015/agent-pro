## Preconditions
- The mock config contains legacy codex events (`item.started`, `item.completed` with `item` field).

## Steps
1. Run fake Codex with the legacy mock config.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
    writeMockConfig(t, req, `{"version":"agent-pro.fake-runner.v1","runner":"fake-codex","stdout_events":[{"type":"item.completed","item":{"id":"m1","type":"message","text":"legacy format ok","status":"completed"}}]}`)
    return nil
}
```
