## Expected

- Exit code 0.
- Dry-run plan text present (would open iTerm / follow-up).
- Not the empty-runner flag error.
- No kill log; no iTerm script content; no agent-run meta.

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
	combined := combinedOut(resp)
	assertTakeoverActionImplemented(t, combined)
	lower := strings.ToLower(combined)
	if strings.Contains(lower, "requires --grok") || strings.Contains(lower, "requires --codex") {
		t.Fatalf("auto-detect dry-run should not demand runner flags, got:\n%s", combined)
	}
	assertExitCode(t, resp, 0)
	if !strings.Contains(lower, "dry-run") && !strings.Contains(lower, "dry run") {
		t.Fatalf("expected dry-run plan text, got:\n%s", combined)
	}
	assertContainsAny(t, combined, "iterm", "iTerm", "open")
	assertNoKillLog(t, req)
	assertNoItermScript(t, req)
	if ids := listAgentSessionIDs(t, req.Home); len(ids) > 0 {
		t.Fatalf("dry-run must not create meta; sessions=%v", ids)
	}
}
```
