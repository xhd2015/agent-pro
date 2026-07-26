---
label: e2e
---

## Expected

- Stderr contains a line matching `agent-run web listening at http://127.0.0.1:<port>` that ends with a newline (not glued to the next character on the same line).

```go
import (
	"regexp"
	"strings"
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	stderr := webProcessStderrText(req)
	re := regexp.MustCompile(`agent-run web listening at https?://127\.0\.0\.1:\d+\n`)
	if !re.MatchString(stderr) {
		t.Fatalf("expected listen URL immediately followed by newline in stderr:\n%q", stderr)
	}
	for _, line := range strings.Split(stderr, "\n") {
		if strings.Contains(line, "listening at http") && strings.Contains(line, "💬") {
			t.Fatalf("listen URL glued to agent output on one line: %q", line)
		}
	}
}
```