---
label: unit
explanation: "sole private type missing_scope hard-fails with needed= and see: topic in stderr"
---

## Expected

- Exit code 1.
- Stdout empty (no partial public listing when private is the only type).
- Stderr contains hard-fail form with needed scope and help topic pointer:
  `channels failed: missing_scope (needed groups:read); see: slack-msg --help --topic add-missing-scope`
- Must not soft-warn (`warning: skipped private channels` absent).

## Exit Code

1

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
	assertExitCode(t, resp, 1)
	if strings.TrimSpace(resp.Stdout) != "" {
		t.Fatalf("expected empty stdout on hard fail, got:\n%s", resp.Stdout)
	}
	assertStderrContains(t, resp, "channels failed: missing_scope (needed groups:read); see: slack-msg --help --topic add-missing-scope")
	if strings.Contains(resp.Stderr, "warning: skipped private channels") {
		t.Fatalf("sole-type missing_scope must hard-fail, not soft-warn:\n%s", resp.Stderr)
	}
}
```
