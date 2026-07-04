## Expected

- Detached run succeeds and registry contains `resp.SessionID`.
- Serve child process command line includes `__serve_sleep_300__`.
- Serve child command line does **not** use the legacy bare token `__serve__`.
- Reexec dispatch prefix/suffix: token matches `__serve_*__` with non-empty slug.

## Side Effects

- Detached session remains until harness cleanup or `kill`.

## Exit Code

- N/A (harness phase).

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.SessionID == "" {
		t.Fatal("expected session id from registry")
	}
	assertRegistryHasSession(t, req.TTYWatchHome, resp.SessionID)

	psLine, err := processArgsLine(t, req.TTYWatchHome, resp.SessionID)
	if err != nil {
		t.Fatal(err)
	}
	wantToken := "__serve_sleep_300__"
	if !strings.Contains(psLine, wantToken) {
		t.Fatalf("serve child ps line missing %q: %q", wantToken, psLine)
	}
	if strings.Contains(psLine, "__serve__") && !strings.Contains(psLine, wantToken) {
		t.Fatalf("legacy __serve__ token still used in ps line: %q", psLine)
	}
	if strings.Count(psLine, "__serve_") < 1 || !strings.Contains(psLine, "__") {
		t.Fatalf("ps line missing __serve_{slug}__ pattern: %q", psLine)
	}
}
```