---
label: e2e
---

## Expected Output

Stderr should start immediately with the token line (no leading blank line):

```text
agent-run web token: <hex>
agent-run web listening at http://127.0.0.1:<port>
```

## Expected

- Stderr does not begin with `\n`.
- First bytes of stderr are `agent-run web token:`.

```go
import (
	"strings"
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	stderr := webProcessStderrText(req)
	if strings.HasPrefix(stderr, "\n") {
		t.Fatalf("stderr must not start with blank line, got:\n%q", stderr)
	}
	const prefix = "agent-run web token:"
	if !strings.HasPrefix(stderr, prefix) {
		t.Fatalf("stderr must start with %q, got:\n%q", prefix, stderr)
	}
}
```