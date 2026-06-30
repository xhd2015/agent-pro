## Expected

- TCP listener becomes ready within timeout.
- `GET /api/terminal/sessions` returns HTTP 200.

```go
import (
	"net/http"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.HTTPStatus != http.StatusOK {
		t.Fatalf("HTTP status: got %d body %q", resp.HTTPStatus, resp.HTTPBody)
	}
}
```