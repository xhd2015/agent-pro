---
label: e2e
---

## Expected

- Exit code 0.
- Stdout indicates no orphans / no matching serves (wording flexible).
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
	if err != nil {
		t.Fatal(err)
	}
	assertSuccess(t, resp)
	lower := strings.ToLower(resp.Stdout)
	ok := strings.Contains(lower, "no orphan") ||
		strings.Contains(lower, "no matching") ||
		strings.Contains(lower, "none") ||
		strings.Contains(lower, "0 serve") ||
		strings.Contains(lower, "no serve")
	if !ok {
		t.Fatalf("expected no-match / no-orphans message; stdout:\n%s", resp.Stdout)
	}
	assertTrailingNewline(t, resp.Stdout, "kill-orphans dry-run-none stdout")
}
```
