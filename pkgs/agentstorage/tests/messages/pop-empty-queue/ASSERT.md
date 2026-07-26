## Expected

- `PopMessages` succeeds with no error.
- Returned slice is empty (length 0).

```go
import "testing"

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertNoError(t, err)
	if resp.Messages == nil {
		t.Fatal("expected non-nil empty messages slice")
	}
	if len(resp.Messages) != 0 {
		t.Fatalf("expected 0 messages, got %d", len(resp.Messages))
	}
}
```