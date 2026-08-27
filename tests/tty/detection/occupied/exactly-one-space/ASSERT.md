## Expected

- ExactlyOneMoreSpace is true.

```go
import "testing"

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	t.Helper()
	_ = d
	_ = req
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if !resp.ExactlyOneMoreSpace {
		t.Fatal("want ExactlyOneMoreSpace=true for draft+space")
	}
}
```
