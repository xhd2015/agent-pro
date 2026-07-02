## Expected

- Terminal status returns HTTP 200.
- `available` is true while the tty turn is still running.
- Response preserves the web chat `session_id`.
- Response exposes the live PTY `terminal_session_id`.

## Errors

- None from `Run`.

## Exit Code

- Test process exits non-zero until a running web-created tty session stores or
  resolves its terminal mapping before assistant completion.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	assertHTTPStatus(t, resp, 200)
	obj := decodeJSONBody(t, resp.HTTPBody)
	if !boolField(obj, "available") {
		t.Fatalf("running tty terminal should be available before response finishes; registry=%q body=%s", req.TerminalSessionID, resp.HTTPBody)
	}
	if stringField(obj, "session_id") != req.ChatSessionID {
		t.Fatalf("terminal status should keep web chat session id %q: %s", req.ChatSessionID, resp.HTTPBody)
	}
	if stringField(obj, "terminal_session_id") != req.TerminalSessionID {
		t.Fatalf("terminal status should expose live PTY id %q while running: %s", req.TerminalSessionID, resp.HTTPBody)
	}
}
```
