## Preconditions
- The mock config uses `stdout_events` with neutral AgentEvent format (no `llm_events`).
- Backward compatibility: agent event format in `stdout_events` must still convert.

## Steps
1. Run fake codex with the mock config.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
    writeMockConfig(t, req, `{"version":"agent-pro.fake-runner.v1","runner":"fake-codex","stdout_events":[{"type":"think","text":"agent format backcompat"}]}`)
    return nil
}
```
