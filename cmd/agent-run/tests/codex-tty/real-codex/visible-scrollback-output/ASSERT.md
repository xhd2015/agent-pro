---
label: codex
explanation: Requires real codex CLI on PATH; verifies real interactive scrollback capture.
---

## Expected

- Exit code 0.
- Stderr contains the prefixed `codex-tty: session-N` registry id.
- Stdout or persisted events contain non-empty visible output.

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
	if strings.Contains(strings.ToLower(resp.Stderr), "banner not detected") {
		t.Fatalf("codex banner not detected:\n%s", resp.Stderr)
	}
	assertSuccess(t, resp)

	if _, ok := parseCodexTTYSessionID(resp.Stderr); !ok {
		t.Fatalf("stderr missing codex-tty session id; stderr:\n%s", resp.Stderr)
	}

	stdout := strings.TrimSpace(resp.Stdout)
	if stdout == "" {
		_, lines := findCodexTTYEventsJSONL(t, req.Home)
		if strings.TrimSpace(strings.Join(lines, "\n")) == "" {
			t.Fatalf("expected non-empty stdout or events from scrollback capture; stderr:\n%s", resp.Stderr)
		}
	}
}
```
