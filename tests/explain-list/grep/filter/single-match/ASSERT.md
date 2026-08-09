---
label: e2e
---

## Expected

- Exit 0.
- Title includes `1 shown of 1` and `limit 10`.
- `marker-k8s` present; `marker-docker` and `marker-redis` absent.
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
	assertNotContains(t, resp.Stderr, "FAKE_AGENT_INVOKED")

	if !strings.Contains(resp.Stdout, "1 shown of 1") || !strings.Contains(resp.Stdout, "limit 10") {
		t.Fatalf("title must use match totals (1 of 1, limit 10):\n%s", resp.Stdout)
	}
	assertContains(t, resp.Stdout, "marker-k8s")
	assertNotContains(t, resp.Stdout, "marker-docker")
	assertNotContains(t, resp.Stdout, "marker-redis")
	assertNotContains(t, resp.Stdout, "No matching explain sessions")
	assertNotContains(t, resp.Stdout, "No explain sessions yet")
}
```
