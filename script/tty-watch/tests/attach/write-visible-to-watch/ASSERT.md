## Expected

- Attach handshake reports `attach_role` attacher.
- Watch stdout contains the attach-written marker.
- Attach output also contains the marker (echo/broadcast).

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
		marker = "ATTACH_WRITE_MARKER_A"
	}
	if !strings.Contains(resp.AttachOutput, `"attach_role":"attacher"`) {
		t.Fatalf("attach write requires attach_mode=attach attacher, handshake %q", resp.AttachOutput)
	}
	if !strings.Contains(resp.WatchOutput, marker) {
		t.Fatalf("watch missing attach write marker %q, got %q", marker, resp.WatchOutput)
	}
	if !strings.Contains(resp.AttachOutput, marker) {
		t.Fatalf("attach missing own write marker %q, got %q", marker, resp.AttachOutput)
	}
}
```