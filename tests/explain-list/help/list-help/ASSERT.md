## Expected

- Exit code 0 (help is not an error).
- Combined stdout+stderr mentions `--limit` and `--color`.
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
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	// Help should succeed. If implementer uses a path that exits non-zero, still
	// require the flag names to appear — but lock exit 0 per common CLI practice.
	assertExitCode(t, resp, 0)
	assertNotContains(t, resp.Stderr, "FAKE_AGENT_INVOKED")

	combined := resp.Stdout + resp.Stderr
	if !strings.Contains(combined, "--limit") {
		t.Fatalf("list help must mention --limit:\n%s", combined)
	}
	if !strings.Contains(combined, "--color") {
		t.Fatalf("list help must mention --color:\n%s", combined)
	}
	if !strings.Contains(strings.ToLower(combined), "list") {
		t.Fatalf("list help should mention list:\n%s", combined)
	}
}
```
