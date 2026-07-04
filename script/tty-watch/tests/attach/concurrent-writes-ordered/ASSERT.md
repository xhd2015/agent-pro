## Expected

- Both attach handshakes report `attach_role` attacher.
- Both complete end-markers appear in combined output.
- No interleaved garbage (partial markers merged mid-string).

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
		markerA = "CONCURRENT_MARKER_A_END"
	}
	markerB := req.AttachInputB
	if markerB == "" {
		markerB = "CONCURRENT_MARKER_B_END"
	}
	combined := resp.Combined
	if combined == "" {
		combined = resp.AttachOutput + resp.AttachBOutput
	}
	if !strings.Contains(combined, markerA) {
		t.Fatalf("missing complete marker A %q in %q", markerA, combined)
	}
	if !strings.Contains(combined, markerB) {
		t.Fatalf("missing complete marker B %q in %q", markerB, combined)
	}
	// Reject obvious mid-marker splice garbage from unsynchronized PTY writes.
	badFragments := []string{
		"CONCURRENT_MARKER_AB_END",
		"CONCURRENT_MARKER_BA_END",
		"CONCURRENT_MARKER__END",
	}
	for _, frag := range badFragments {
		if strings.Contains(combined, frag) {
			t.Fatalf("interleaved garbage fragment %q in output %q", frag, combined)
		}
	}
}
```