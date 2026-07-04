## Expected

- Exit code 0 from `list`.
- `resp.SessionID` is `test-with-grok` (verbatim custom id, no forced `session-` prefix).
- Registry file `test-with-grok.json` exists.
- `resp.ListOutput` contains `test-with-grok` and `sleep`.
- Session TCP endpoint is reachable.

## Exit Code

0

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.ExitCode != 0 {
		t.Fatalf("list exit %d, output %q", resp.ExitCode, resp.ListOutput)
	}
	if resp.SessionID != "test-with-grok" {
		t.Fatalf("expected custom session id test-with-grok, got %q", resp.SessionID)
	}
	if !resp.RegistryExists {
		t.Fatalf("registry should exist for %s", resp.SessionID)
	}
	assertRegistryHasSession(t, req.TTYWatchHome, "test-with-grok")
	if !resp.SessionRunning {
		t.Fatalf("session %s should be reachable after detach", resp.SessionID)
	}
	if !strings.Contains(resp.ListOutput, "test-with-grok") {
		t.Fatalf("list missing custom id: %q", resp.ListOutput)
	}
	if !strings.Contains(resp.ListOutput, "sleep") {
		t.Fatalf("list missing command sleep: %q", resp.ListOutput)
	}
}
```