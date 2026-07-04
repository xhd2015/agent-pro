## Expected

- Attach handshake reports `attach_role` attacher while screen writer is connected.
- Run writer stdout contains the attach-written marker.
- Watch stdout contains the same marker.
- Screen writer and attach can coexist (multi-writer attach mode).

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
		t.Fatalf("attach requires attacher role while screen writer connected, handshake %q", resp.AttachOutput)
	}
	if !strings.Contains(resp.RunOutput, marker) {
		t.Fatalf("run writer missing attach marker %q, got %q", marker, resp.RunOutput)
	}
	if !strings.Contains(resp.WatchOutput, marker) {
		t.Fatalf("watch missing attach marker %q, got %q", marker, resp.WatchOutput)
	}
}
```