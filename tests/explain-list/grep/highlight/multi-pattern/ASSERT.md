---
label: e2e
---

## Expected

- Exit 0.
- Session listed (`1 shown of 1`).
- Bold-red wrap of `docker` and of `kubernetes` (original lowercasing in body).
- Trailing newline.

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
	assertHasANSI(t, resp.Stdout)

	if !strings.Contains(resp.Stdout, "1 shown of 1") {
		t.Fatalf("expected matching session listed:\n%s", resp.Stdout)
	}

	assertContains(t, resp.Stdout, "\x1b[1;31mdocker\x1b[0m")
	assertContains(t, resp.Stdout, "\x1b[1;31mkubernetes\x1b[0m")
}
```
