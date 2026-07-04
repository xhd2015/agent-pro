## Expected

- Registry contains at least one `session-N` entry after default run starts.
- `resp.SessionID` matches a registry file.
- `resp.RegistryExists` is true.

## Side Effects

- Detached session remains until harness cleanup or `kill`.

## Exit Code

- N/A (harness phase, no direct CLI exit assertion).

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
	if !strings.HasPrefix(resp.SessionID, "session-") {
		t.Fatalf("session id should be session-N, got %q", resp.SessionID)
	}
	if !resp.RegistryExists {
		t.Fatalf("registry should exist for %s", resp.SessionID)
	}
	assertRegistryHasSession(t, req.TTYWatchHome, resp.SessionID)
	if len(resp.RegistryIDs) == 0 {
		t.Fatal("expected at least one registry id")
	}
}
```