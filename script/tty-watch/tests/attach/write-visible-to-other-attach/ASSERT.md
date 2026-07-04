## Expected

- Both attach handshakes report `attach_role` attacher (multi-writer attach mode).
- Attacher B output contains the marker written only by attacher A.
- Attacher A output contains the marker.

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
		t.Fatalf("attach A requires attacher role, handshake %q", resp.AttachOutput)
	}
	if !strings.Contains(resp.AttachBOutput, `"attach_role":"attacher"`) {
		t.Fatalf("attach B requires attacher role, handshake %q", resp.AttachBOutput)
	}
	if !strings.Contains(resp.AttachBOutput, marker) {
		t.Fatalf("attach B missing A's marker %q, got %q", marker, resp.AttachBOutput)
	}
	if !strings.Contains(resp.AttachOutput, marker) {
		t.Fatalf("attach A missing own marker %q, got %q", marker, resp.AttachOutput)
	}
}
```