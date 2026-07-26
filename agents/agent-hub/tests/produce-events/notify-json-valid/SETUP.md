## Preconditions
- A valid NormalizedEvent JSON is provided via --json.

## Steps
1. Run `agent-hub notify --json '{"event_type":"agent.session.started","runner":"fake-opencode","runner_session_id":"s1"}'`
2. Then fetch with --consumer-id test --limit 1 to verify persistence.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
    req.Args = []string{"notify", "--json", `{"event_type":"agent.session.started","runner":"fake-opencode","runner_session_id":"s1"}`}
    return nil
}
```
