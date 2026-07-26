## Expected

- Stale grok entry removed.
- Live codex entry returned.

```go
import (
	"os"
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertNoError(t, err)
	if resp.LookupRunnerID != "codex-tty" {
		t.Fatalf("expected codex-tty after stale grok skip, got %q", resp.LookupRunnerID)
	}
	stalePath := registryPathFor(req.Home, "grok-tty-registry", "session-1")
	if _, err := os.Stat(stalePath); err == nil { t.Fatal("expected stale grok registry file removed") }
}
```
