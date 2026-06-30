## Expected

- After PATCH, list entry for session has `Name == after-rename`.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	for _, s := range resp.Sessions {
		if s.ID == resp.SessionID {
			if s.Name != req.RenameTo {
				t.Fatalf("name: got %q want %q", s.Name, req.RenameTo)
			}
			return
		}
	}
	t.Fatalf("session %q not in list", resp.SessionID)
}
```