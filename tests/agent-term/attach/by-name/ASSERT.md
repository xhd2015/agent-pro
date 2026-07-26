## Expected

- WS attach probe by name succeeds (no error, exit 0).

```go
import "testing"

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.ExitCode != 0 {
		t.Fatalf("attach probe failed: %s", resp.Stderr)
	}
}
```