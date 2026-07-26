## Preconditions
- The mock config uses `stdout_events` (no `llm_events`).
- The deprecated field should still work but emit a stderr warning.

## Steps
1. Run fake codex with the mock config.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
    writeMockConfig(t, req, `{"version":"agent-pro.fake-runner.v1","runner":"fake-codex","stdout_events":[{"type":"think","text":"deprecated codex works"}]}`)
    return nil
}
```
