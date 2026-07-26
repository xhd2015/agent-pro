## Preconditions
- The mock config sets `delay_ms` to zero.

## Steps
1. Run fake Codex with two events.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
    writeMockConfig(t, req, `{"version":"agent-pro.fake-runner.v1","runner":"fake-codex","delay_ms":0,"llm_events":[{"type":"message","text":"fast one"},{"type":"message","text":"fast two"}]}`)
    return nil
}
```

