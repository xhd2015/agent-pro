## Expected
- Output[0]: non-empty (PhaseEnd for m1 formatted).
- Output[1]: non-empty (think event formatted).
- Output[2]: non-empty (PhaseEnd for m2 formatted — state was reset by think).

```go
import (
	"strings"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if len(resp.Output) != 3 {
		t.Fatalf("expected 3 outputs, got %d", len(resp.Output))
	}
	if resp.Output[0] == "" {
		t.Fatalf("first PhaseEnd must produce formatted output")
	}
	if !strings.Contains(resp.Output[0], "first") {
		t.Fatalf("first PhaseEnd output must contain text 'first', got: %q", resp.Output[0])
	}
	if resp.Output[1] == "" {
		t.Fatalf("think event must always produce formatted output")
	}
	if !strings.Contains(resp.Output[1], "hmm") {
		t.Fatalf("think output must contain 'hmm', got: %q", resp.Output[1])
	}
	if resp.Output[2] == "" {
		t.Fatalf("PhaseEnd after think (state reset) must produce formatted output")
	}
	if !strings.Contains(resp.Output[2], "second") {
		t.Fatalf("second PhaseEnd output must contain text 'second', got: %q", resp.Output[2])
	}
}
```
