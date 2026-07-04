## Expected

- Attach exits 0 after Ctrl-] detach.
- Registry entry survives detach.
- Session remains reachable; `list` includes session id.

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
		t.Fatalf("attach detach expected exit 0, got %d output %q", resp.ExitCode, resp.AttachOutput)
	}
	if resp.SessionID == "" {
		t.Fatal("expected detached session id")
	}
	if !resp.RegistryExists {
		t.Fatalf("registry should survive attach detach for %s", resp.SessionID)
	}
	if !resp.SessionRunning {
		t.Fatalf("session %s should still be reachable after attach detach", resp.SessionID)
	}
	if !strings.Contains(resp.ListOutput, resp.SessionID) {
		t.Fatalf("list should include %s, got %q", resp.SessionID, resp.ListOutput)
	}
}
```