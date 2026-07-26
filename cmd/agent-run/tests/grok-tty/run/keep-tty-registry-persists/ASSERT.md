---
label: e2e
---

## Expected

- Exit code 0 (run completes normally).
- Registry file exists at `AGENT_RUN_HOME/grok-tty-registry/<session-id>.json` after run exits.
- The session ID is printed on stderr as `grok-tty: session-N`.

```go
import (
	"testing"
	"os"
	"path/filepath"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertSuccess(t, resp)

	sessionID, ok := parseGrokTTYSessionID(resp.Stderr)
	if !ok {
		t.Fatalf("missing grok-tty session id in stderr:\n%s", resp.Stderr)
	}

	regPath := filepath.Join(grokTTYRegistryDir(req.Home), sessionID+".json")
	if _, statErr := os.Stat(regPath); statErr != nil {
		t.Fatalf("registry entry not persisted at %s after --keep-tty run: %v\nstderr:\n%s", regPath, statErr, resp.Stderr)
	}
}
```
