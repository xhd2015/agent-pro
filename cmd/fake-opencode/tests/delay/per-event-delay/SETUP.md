## Preconditions
- A message event carries `delay_ms: 2000`.

## Steps
1. Write mock config with a message event that has a 2s pre-emission delay.
2. Run fake-opencode. The overridden Run measures elapsed wall time.

```go
import (
    "bytes"
    "context"
    "errors"
    "testing"
    "time"
    "os/exec"
    "os"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
    req.Operation = "per_event_delay"
    writeMockConfig(t, req, `{"version":"agent-pro.fake-runner.v1","runner":"fake-opencode","session_id":"sess_delay","llm_events":[{"type":"message","text":"delayed","delay_ms":2000}]}`)
    return nil
}
```
