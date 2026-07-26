## Expected

- Input bytes sent over WS appear in captured output.

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(resp.WSOutput, "roundtrip-marker") {
		t.Fatalf("expected roundtrip-marker in output, got: %q", resp.WSOutput)
	}
}
```