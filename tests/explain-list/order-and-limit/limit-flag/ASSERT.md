---
label: e2e
---

## Expected

- Exit 0.
- Title includes `3 shown of 5` and `limit 3`.
- `question-04`, `question-03`, `question-02` present; `question-00`,
  `question-01` absent.
- Newest (`question-04`) listed before `question-03`.
- Trailing newline; no ANSI.

## Side Effects

- None.

## Errors

- None.

## Exit Code

- 0.

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
	assertExitCode(t, resp, 0)
	assertStdoutEndsWithNewline(t, resp.Stdout)
	assertNoANSI(t, resp.Stdout)

	if !strings.Contains(resp.Stdout, "3 shown of 5") || !strings.Contains(resp.Stdout, "limit 3") {
		t.Fatalf("title must include 3 shown and limit 3:\n%s", resp.Stdout)
	}
	assertContains(t, resp.Stdout, "question-04")
	assertContains(t, resp.Stdout, "question-03")
	assertContains(t, resp.Stdout, "question-02")
	assertNotContains(t, resp.Stdout, "question-00")
	assertNotContains(t, resp.Stdout, "question-01")

	i4 := strings.Index(resp.Stdout, "question-04")
	i3 := strings.Index(resp.Stdout, "question-03")
	if i4 < 0 || i3 < 0 || i4 > i3 {
		t.Fatalf("expected question-04 before question-03; i4=%d i3=%d\n%s", i4, i3, resp.Stdout)
	}
}
```
