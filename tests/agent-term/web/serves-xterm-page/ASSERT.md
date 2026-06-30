## Expected

- HTTP GET returns status 200.
- Body contains xterm.js markers (`xterm` or embedded terminal markup).

```go
import (
	"net/http"
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.HTTPStatus != http.StatusOK {
		t.Fatalf("status %d", resp.HTTPStatus)
	}
	lower := strings.ToLower(resp.HTTPBody)
	if !strings.Contains(lower, "xterm") {
		t.Fatalf("expected xterm markup in page body")
	}
}
```