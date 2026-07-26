## Expected
- Output[0] must be non-empty (formatted output present).
- Output[0] must contain the text "done".

```go
import (
	"strings"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if len(resp.Output) != 1 {
		t.Fatalf("expected 1 output, got %d", len(resp.Output))
	}
	if resp.Output[0] == "" {
		t.Fatalf("standalone PhaseEnd must produce formatted output")
	}
	if !strings.Contains(resp.Output[0], "done") {
		t.Fatalf("output must contain message text, got: %q", resp.Output[0])
	}
}
```
