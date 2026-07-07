## Expected

- Exactly one registry entry remains for codex-tty.
- Registry id equals original `terminal_session_id` (no `session-2` allocation).

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	ids := listRegistryIDs(t, req.Home, req.Runner)
	if len(ids) != 1 {
		t.Fatalf("expected one registry entry after follow-up, got %v", ids)
	}
	if ids[0] != req.TerminalSessionID {
		t.Fatalf("follow-up changed registry id: want %q got %q", req.TerminalSessionID, ids[0])
	}
	_, body := getSessionDetail(t, req, req.Runner, req.ChatSessionID)
	if terminalSessionIDFromDetail(body) != req.TerminalSessionID {
		t.Fatalf("session detail terminal_session_id changed after follow-up: %s", body)
	}
}
```
