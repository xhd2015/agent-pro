## Preconditions
- The mock config contains both `llm_events` and `stdout_events` with different content.
- `llm_events` should take precedence.

## Steps
1. Run fake opencode with the mock config.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
    writeMockConfig(t, req, `{"version":"agent-pro.fake-runner.v1","runner":"fake-opencode","session_id":"sess_prec","llm_events":[{"type":"think","text":"from llm_events"}],"stdout_events":[{"type":"think","text":"from stdout_events"}]}`)
    return nil
}
```
