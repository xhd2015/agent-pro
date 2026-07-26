---
label: e2e
---

## Expected

- Exit 0.
- Title includes shown count 10 and limit 10 (e.g. `10 shown of 12, limit 10`).
- Newest questions `question-11` … `question-02` appear; oldest `question-00`
  and `question-01` do not.
- First card is newest (`question-11`).
- Trailing newline; no ANSI.

## Side Effects

- None.

## Errors

- None.

## Exit Code

- 0.

```go
import (
	"fmt"
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

	// Title must report shown + limit.
	if !strings.Contains(resp.Stdout, "10 shown of 12") || !strings.Contains(resp.Stdout, "limit 10") {
		t.Fatalf("title must include 10 shown and limit 10:\n%s", resp.Stdout)
	}

	// Newest 10: indices 2..11 inclusive.
	for i := 2; i <= 11; i++ {
		assertContains(t, resp.Stdout, fmt.Sprintf("question-%02d", i))
	}
	// Oldest 2 excluded by default limit.
	assertNotContains(t, resp.Stdout, "question-00")
	assertNotContains(t, resp.Stdout, "question-01")

	// Newest first: question-11 appears before question-10.
	i11 := strings.Index(resp.Stdout, "question-11")
	i10 := strings.Index(resp.Stdout, "question-10")
	if i11 < 0 || i10 < 0 || i11 > i10 {
		t.Fatalf("expected question-11 before question-10; i11=%d i10=%d\n%s", i11, i10, resp.Stdout)
	}
}
```
