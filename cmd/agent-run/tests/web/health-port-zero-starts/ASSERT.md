## Expected

- Server starts with OS-assigned port (`--port 0`).
- `req.WebBaseURL` is non-empty.
- Health endpoint returns HTTP 200 with valid Bearer.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if req.WebBaseURL == "" {
		t.Fatal("expected WebBaseURL to be set after --port 0 startup")
	}
	if resp.HTTPStatus != 200 {
		t.Fatalf("expected HTTP 200 from health on dynamic port, got %d body=%q", resp.HTTPStatus, resp.HTTPBody)
	}
}
```