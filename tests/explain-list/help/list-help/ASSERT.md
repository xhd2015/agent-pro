---
label: e2e
---

## Expected

- Exit code 0 (help is not an error).
- Combined stdout+stderr mentions `--limit`, `--grep`, `--or`, `--and`, and `--color`.
- Mentions `list` (subcommand context).
- Fake agent not invoked.

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
	// Help should succeed. If implementer uses a path that exits non-zero, still
	// require the flag names to appear — but lock exit 0 per common CLI practice.
	assertExitCode(t, resp, 0)
	assertNotContains(t, resp.Stderr, "FAKE_AGENT_INVOKED")

	combined := resp.Stdout + resp.Stderr
	for _, flag := range []string{"--limit", "--grep", "--or", "--and", "--color"} {
		if !strings.Contains(combined, flag) {
			t.Fatalf("list help must mention %s:\n%s", flag, combined)
		}
	}
	if !strings.Contains(strings.ToLower(combined), "list") {
		t.Fatalf("list help should mention list:\n%s", combined)
	}
}
```