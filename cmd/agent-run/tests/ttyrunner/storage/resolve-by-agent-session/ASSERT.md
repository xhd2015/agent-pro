## Expected

- Resolver returns live registry for agent session id.
- `RunnerID` is `grok-tty`, `TCPReachable` true.

```go
import (
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertNoError(t, err)
	if resp.TTYSession == nil { t.Fatal("expected TTYSession") }
	s := resp.TTYSession
	if s.TerminalSessionID != "session-1" { t.Fatalf("terminal_session_id: got %q", s.TerminalSessionID) }
	if s.RunnerID != "grok-tty" { t.Fatalf("runner_id: got %q", s.RunnerID) }
	if !s.TCPReachable { t.Fatal("expected TCP reachable") }
}
```
