## Expected

- The generated session detail reaches `status:"finished"`.
- The generated session detail includes a non-empty `terminal_session_id`.
- Terminal status returns HTTP 200.
- Terminal status reports `available:true` and echoes the generated
  `terminal_session_id`.

## Side Effects

- Creates one isolated web codex-tty session under the test `AGENT_RUN_HOME`.

## Errors

- None from `Run`.

## Exit Code

- Test process exits non-zero while web-created codex-tty sessions tear down the
  PTY server before the frontend can attach.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if req.TerminalSessionID == "" {
		t.Fatalf("generated session did not persist terminal_session_id")
	}
	assertHTTPStatus(t, resp, 200)
	requireTerminalMappingAvailable(t, req, resp.HTTPBody)
}
```
