## Expected

- First connection sees `scrollback-marker`.
- Reconnect output also contains `scrollback-marker` (scrollback replay).

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(resp.WSOutput, "scrollback-marker") {
		t.Fatalf("first attach missing marker: %q", resp.WSOutput)
	}
	if !strings.Contains(resp.ReconnectOutput, "scrollback-marker") {
		t.Fatalf("reconnect missing scrollback marker: %q", resp.ReconnectOutput)
	}
}
```