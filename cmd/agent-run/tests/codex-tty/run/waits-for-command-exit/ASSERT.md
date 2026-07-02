## Expected

- Exit code 0.
- Run does not return before fake TUI completes (captured response present).
- Stderr contains `codex-tty: session-` prefix line.

## Exit Code

0

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	assertSuccess(t, resp)
	if _, ok := parseCodexTTYSessionID(resp.Stderr); !ok {
		t.Fatalf("expected codex-tty session id on stderr:\n%s", resp.Stderr)
	}
	if !strings.Contains(resp.Stdout, "ping") && !strings.Contains(resp.Stdout, "Response:") {
		t.Fatalf("expected captured TUI output after command exit, stdout:\n%s", resp.Stdout)
	}
}
```