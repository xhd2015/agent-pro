## Expected

- Non-zero exit.
- Stderr contains `Error:` and `parse config.json` (or equivalent parse failure).

```go
import (
	"strings"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	_ = req
	if err != nil {
		t.Fatal(err)
	}
	if resp.ExitCode == 0 {
		t.Fatal("expected non-zero exit for corrupt config")
	}
	assertContains(t, resp.Stderr, "Error:")
	if !strings.Contains(resp.Stderr, "parse config.json") && !strings.Contains(resp.Stderr, "invalid") {
		t.Fatalf("expected parse failure in stderr:\n%s", resp.Stderr)
	}
}
```
