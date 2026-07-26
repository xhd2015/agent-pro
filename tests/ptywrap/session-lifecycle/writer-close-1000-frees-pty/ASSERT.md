## Expected

- After the writer closes the WebSocket with code **1000**, the session child
  process is **not** still running (`ProcessAlive == false`).
- This frees the OS PTY slot. Session metadata may remain for scrollback, but
  must not pin a live process.

```go
import "testing"

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.ProcessAlive {
		t.Fatalf("writer close 1000 left session child running (PTY leak): session=%s sessions=%d",
			resp.SessionID, resp.SessionCount)
	}
}
```
