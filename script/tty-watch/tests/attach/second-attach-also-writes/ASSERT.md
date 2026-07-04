## Expected

- Both attach handshakes report `attach_role` attacher.
- Both distinct markers appear in combined attach output.
- Neither marker is dropped when a second attacher writes.

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(resp.AttachOutput, `"attach_role":"attacher"`) {
		t.Fatalf("attach A requires attacher role, handshake %q", resp.AttachOutput)
	}
	if !strings.Contains(resp.AttachBOutput, `"attach_role":"attacher"`) {
		t.Fatalf("attach B requires attacher role, handshake %q", resp.AttachBOutput)
	}
	markerA := req.AttachInput
	if markerA == "" {
		markerA = "ATTACH_WRITE_MARKER_A"
	}
	markerB := req.AttachInputB
	if markerB == "" {
		markerB = "ATTACH_WRITE_MARKER_B"
	}
	combined := resp.Combined
	if combined == "" {
		combined = resp.AttachOutput + resp.AttachBOutput
	}
	if !strings.Contains(combined, markerA) {
		t.Fatalf("missing marker A %q in combined output %q", markerA, combined)
	}
	if !strings.Contains(combined, markerB) {
		t.Fatalf("missing marker B %q in combined output %q", markerB, combined)
	}
}
```