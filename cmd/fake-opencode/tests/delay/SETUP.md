## Preconditions
- The mock config uses `llm_events` with the neutral AgentEvent format.
- A `sleep` event type with `delay_ms` produces no output and pauses event emission.
- Individual events may carry a `delay_ms` field for a pre-emission delay.

## Steps
1. Inherit the fake-opencode binary and environment from the root SETUP.md.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
    t.Log("setting up delay test group")
    return nil
}
```
