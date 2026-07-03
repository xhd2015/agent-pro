## Expected

- `ResolveByTerminalID` returns enriched session.
- `AgentSessionID` and `TTY` snapshot populated from `tty.json`.

```go
import (
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertNoError(t, err)
	if resp.TTYSession == nil { t.Fatal("expected TTYSession") }
	s := resp.TTYSession
	if s.AgentSessionID != "sess_terminal_lookup" { t.Fatalf("agent_session_id: got %q", s.AgentSessionID) }
	if s.TTY == nil { t.Fatal("expected TTY snapshot from tty.json") }
}
```
