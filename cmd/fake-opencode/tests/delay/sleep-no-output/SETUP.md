## Preconditions
- Mock config has a sleep event followed by a single message.

## Steps
1. Write mock config with sleep(100ms) then message "only".
2. Run fake-opencode.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
    writeMockConfig(t, req, `{"version":"agent-pro.fake-runner.v1","runner":"fake-opencode","session_id":"sess_snoop","llm_events":[{"type":"sleep","delay_ms":100},{"type":"message","text":"only"}]}`)
    return nil
}
```
