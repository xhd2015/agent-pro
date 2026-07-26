## Expected
- gen-commit-msg succeeds and prints the parsed commit message to stdout.

```go
import (
	"strings"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if resp.Err != nil {
		t.Fatalf("gen-commit-msg failed: %v\nstderr:\n%s", resp.Err, resp.Stderr)
	}
	if !strings.Contains(resp.Stdout, "feat: add feature") {
		t.Fatalf("stdout missing title, got:\n%s", resp.Stdout)
	}
	if !strings.Contains(resp.Stdout, "Implement feature X") {
		t.Fatalf("stdout missing description, got:\n%s", resp.Stdout)
	}
}
```