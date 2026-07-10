---
label: unit
explanation: "all types missing_scope -> hard fail exit 1 with see: topic; no partial stdout"
---

## Expected

- Exit code 1.
- Stdout empty.
- Stderr contains `channels failed:` and `missing_scope`.
- Prefer including `needed` when present (`needed channels:read` and/or `needed groups:read`);
  assert at least one `needed` fragment so hard-error form is locked.
- Stderr includes help topic pointer:
  `see: slack-msg --help --topic add-missing-scope`
- Not a soft-only warning path (must hard-fail, not exit 0 with only warnings).

## Exit Code

1

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	assertExitCode(t, resp, 1)
	if strings.TrimSpace(resp.Stdout) != "" {
		t.Fatalf("expected empty stdout when all types fail, got:\n%s", resp.Stdout)
	}
	assertStderrContains(t, resp, "channels failed:")
	assertStderrContains(t, resp, "missing_scope")
	if !strings.Contains(resp.Stderr, "needed ") {
		t.Fatalf("hard fail should include needed= when Slack provides it:\n%s", resp.Stderr)
	}
	assertStderrContains(t, resp, "see: slack-msg --help --topic add-missing-scope")
}
```
