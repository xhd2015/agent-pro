---
label: codex
explanation: Requires real codex CLI on PATH; for design verification and debugging.
---

## Expected

- Exit code 0.
- Stderr does not report `banner not detected` or codex TUI banner timeout.
- Stderr contains prefixed session id `codex-tty: session-`.

## Exit Code

0

```go
import (
	"strings"
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	if err != nil {
		t.Fatal(err)
	}
	assertSuccess(t, resp)
	lower := strings.ToLower(resp.Stderr)
	if strings.Contains(lower, "banner not detected") {
		t.Fatalf("real codex banner not detected, stderr:\n%s", resp.Stderr)
	}
	if _, ok := parseCodexTTYSessionID(resp.Stderr); !ok {
		t.Fatalf("expected codex-tty session id on stderr:\n%s", resp.Stderr)
	}
}
```