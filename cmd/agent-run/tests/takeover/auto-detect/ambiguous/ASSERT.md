## Expected

- Exit code non-zero.
- Error indicates ambiguity between Grok and Codex (or both providers / cannot choose).
- Mentions both providers (grok and codex) in some form.
- Not merely the old empty-runner flag error alone (must be an ambiguity-style message once implemented).
- No iTerm script; no kill log; no agent-run meta.

## Exit Code

non-zero

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
	if resp.ExitCode == 0 {
		t.Fatalf("expected non-zero exit for ambiguous auto-detect, got 0\nstdout:\n%s\nstderr:\n%s",
			resp.Stdout, resp.Stderr)
	}
	combined := combinedOut(resp)
	assertTakeoverActionImplemented(t, combined)
	lower := strings.ToLower(combined)
	// Must not stop at "requires --grok/--codex" once auto-detect exists.
	if strings.Contains(lower, "requires --grok") && !strings.Contains(lower, "ambig") {
		t.Fatalf("want ambiguous multi-provider error, still empty-runner flag gate:\n%s", combined)
	}
	assertContainsAny(t, combined,
		"ambiguous",
		"ambiguity",
		"both",
		"multiple",
		"more than one",
		"conflict",
	)
	// Anchor to both provider kinds.
	if !strings.Contains(lower, "grok") {
		t.Fatalf("ambiguous error should mention grok, got:\n%s", combined)
	}
	if !strings.Contains(lower, "codex") {
		t.Fatalf("ambiguous error should mention codex, got:\n%s", combined)
	}
	assertNoItermScript(t, req)
	assertNoKillLog(t, req)
	if ids := listAgentSessionIDs(t, req.Home); len(ids) > 0 {
		t.Fatalf("ambiguous auto-detect must not create meta; sessions=%v", ids)
	}
}
```
