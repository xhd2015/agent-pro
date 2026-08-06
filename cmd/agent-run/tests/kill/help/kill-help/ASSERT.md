## Expected

- Exit code 0.
- Stdout documents `kill` usage including a session-id placeholder and `--dry-run`.
- Stdout ends with trailing newline `\n`.

## Exit Code

0

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
	assertExitCode(t, resp, 0)
	out := resp.Stdout
	assertContains(t, out, "kill")
	assertContains(t, out, "--dry-run")
	// session-id appears as a required positional in usage (various phrasings OK)
	lower := strings.ToLower(out)
	if !strings.Contains(lower, "session") {
		t.Fatalf("kill --help should document session-id; stdout:\n%s", out)
	}
	assertTrailingNewline(t, out, "kill --help stdout")
}
```
