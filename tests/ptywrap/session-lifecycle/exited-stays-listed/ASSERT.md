## Expected

- Session remains in list after command exits.
- `Status` is `"exited"`.

```go
import "testing"

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	for _, s := range resp.Sessions {
		if s.ID == resp.SessionID {
			if s.Status != "exited" {
				t.Fatalf("status: got %q want exited", s.Status)
			}
			return
		}
	}
	t.Fatalf("session %q not found after exit", resp.SessionID)
}
```