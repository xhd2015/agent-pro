## Expected

- `ready=true`, `state=idle`.
- Scrollback contains `git working tree status` (fixture precondition).
- Must **not** return `state=busy` with reason `agent still responding`.

## Exit Code

N/A (direct package call)

```go
import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	text, err := os.ReadFile(filepath.Join(req.TestdataDir, fixtureFalsePositiveSession18))
	if err != nil {
		t.Fatal(err)
	}
	lower := strings.ToLower(string(text))
	if !strings.Contains(lower, "git working tree status") && !strings.Contains(lower, "working tree") {
		t.Fatalf("fixture must contain git working tree scrollback")
	}
	if !strings.Contains(string(text), "❯") {
		t.Fatalf("fixture must contain idle prompt marker")
	}
	assertWritable(t, "session-18 false positive", resp.Status, true, "idle", "")
	if resp.Status.State == "busy" {
		t.Fatalf("false positive: got busy reason=%q", resp.Status.Reason)
	}
}
```