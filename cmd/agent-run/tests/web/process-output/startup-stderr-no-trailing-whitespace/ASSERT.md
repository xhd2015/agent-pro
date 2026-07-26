---
label: e2e
---

## Expected

- Stderr does not end with space or tab.
- Last character of stderr is `\n` (from the listen URL line).

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
	if stderr == "" {
		t.Fatal("stderr is empty")
	}
	last := stderr[len(stderr)-1]
	if last == ' ' || last == '\t' {
		t.Fatalf("stderr must not end with whitespace, last char %q in:\n%q", last, stderr)
	}
	if last != '\n' {
		t.Fatalf("stderr must end with newline from listen line, last char %q in:\n%q", last, stderr)
	}
	if trimmed := strings.TrimRight(stderr, " \t"); trimmed != stderr {
		t.Fatalf("stderr contains trailing space/tab:\n%q", stderr)
	}
}
```