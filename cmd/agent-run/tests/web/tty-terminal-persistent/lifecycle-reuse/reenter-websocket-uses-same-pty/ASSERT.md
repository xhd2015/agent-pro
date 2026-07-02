## Expected

- Both websocket attaches receive the transcript from `session-1`.
- The fake PTY server observes exactly two attach connections to the same
  registry entry.
- No `session-2` registry entry appears.

## Side Effects

- Two websocket connections are opened and closed.

## Errors

- None from `Run`.

## Exit Code

- Test process exits non-zero until websocket proxy resolves through
  `terminal_session_id`.

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.WSError != "" {
		t.Fatalf("websocket attach failed: %s output=%q", resp.WSError, resp.WSOutput)
	}
	if strings.Count(resp.WSOutput, req.RegistryTranscript) != 2 {
		t.Fatalf("expected two attaches to mapped PTY transcript %q, got %q", req.RegistryTranscript, resp.WSOutput)
	}
	if resp.PTYConnectionSeen != 2 {
		t.Fatalf("fake PTY connection count=%d want 2", resp.PTYConnectionSeen)
	}
	requireSameStringSlice(t, registryIDs(t, req), []string{req.TerminalSessionID})
}
```
