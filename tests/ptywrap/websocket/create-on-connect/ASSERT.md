## Expected

- WS connect without `session_id` creates a session.
- Server sends `session_id` JSON before PTY output.
- Listed session has requested name and cwd.

```go
import (
	"testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.SessionID == "" {
		t.Fatal("expected session_id from create-on-connect")
	}
	found := false
	for _, s := range resp.Sessions {
		if s.ID == resp.SessionID {
			found = true
			if s.Name != req.Name {
				t.Fatalf("name: got %q want %q", s.Name, req.Name)
			}
			if s.Cwd != req.Cwd {
				t.Fatalf("cwd: got %q want %q", s.Cwd, req.Cwd)
			}
		}
	}
	if !found {
		t.Fatalf("session %q not found in list", resp.SessionID)
	}
}
```