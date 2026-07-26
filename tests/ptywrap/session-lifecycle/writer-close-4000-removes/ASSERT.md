## Expected

- Writer close code **4000** kills the child (`ProcessAlive == false`).
- Session is removed from GET `/api/terminal/sessions` (`SessionListed == false`).

```go
import "testing"

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.ProcessAlive {
		t.Fatalf("close 4000 left child running: session=%s", resp.SessionID)
	}
	if resp.SessionListed {
		t.Fatalf("close 4000 left session listed: session=%s", resp.SessionID)
	}
}
```
