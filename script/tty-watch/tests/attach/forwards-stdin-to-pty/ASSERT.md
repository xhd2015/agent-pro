## Expected

- Attach stdout contains the unique stdin marker after typing.
- Marker is complete (not truncated).

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	marker := req.AttachInput
	if marker == "" {
		marker = "ATTACH_STDIN_MARKER"
	}
	out := resp.AttachOutput
	if out == "" {
		out = resp.Combined
	}
	if !strings.Contains(out, marker) {
		t.Fatalf("attach stdout missing stdin marker %q, got %q", marker, out)
	}
}
```