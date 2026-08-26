## Expected

- Exit 0.
- Stdout is `{}` (pretty empty object).
- No agent invocation.

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
	if strings.TrimSpace(resp.Stdout) != "{}" {
		t.Fatalf("stdout = %q, want {}", resp.Stdout)
	}
	if strings.Contains(resp.Stderr, "FAKE_AGENT_INVOKED") {
		t.Fatalf("show-config must not invoke agent:\n%s", resp.Stderr)
	}
}
```
