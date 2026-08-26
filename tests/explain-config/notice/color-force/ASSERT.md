## Expected

- Non-zero exit (failing fake).
- Stderr contains gray-wrapped `notice:` then `agent-runner=codex (from config)`.

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
		t.Fatal("expected non-zero exit (failing fake)")
	}
	want := "\x1b[90mnotice:\x1b[0m agent-runner=codex (from config)"
	if !strings.Contains(resp.Stderr, want) {
		t.Fatalf("stderr missing gray notice %q:\n%q", want, resp.Stderr)
	}
}
```
