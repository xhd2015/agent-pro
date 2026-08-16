---
label: unit
explanation: "multi-type list soft-skips private missing_scope; public human lines + warning with see: topic"
---

## Expected Output

```
C0ALE44K5J6  #general  public  member
C0OTHERCHAN  #random  public  -
```

## Expected

- Exit code 0.
- Stdout: non-archived public channels only, name-sorted (no private `agent-pro-debug`).
- Stderr contains soft-fail warning with needed scope and help topic pointer:
  `warning: skipped private channels (missing groups:read); see: slack-msg --help --topic add-missing-scope`
  (order relative to stdout not asserted).
- Archived `#old-stuff` still excluded.
- Trailing newline after last human line.

## Exit Code

0

```go
import (
	"strings"
	"testing"

	"github.com/xhd2015/doctest/assert"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	assertExitCode(t, resp, 0)
	assertStderrContains(t, resp, "warning: skipped private channels (missing groups:read); see: slack-msg --help --topic add-missing-scope")
	if strings.Contains(resp.Stdout, "agent-pro-debug") || strings.Contains(resp.Stdout, "C0AGENTDBG1") {
		t.Fatalf("private channel must not appear after soft-skip:\n%s", resp.Stdout)
	}
	if strings.Contains(resp.Stdout, "old-stuff") || strings.Contains(resp.Stdout, "C0ARCHIVED1") {
		t.Fatalf("archived channel must be excluded:\n%s", resp.Stdout)
	}
	assert.Output(t, resp.Stdout, `C0ALE44K5J6  #general  public  member
C0OTHERCHAN  #random  public  -
`)
}
```
