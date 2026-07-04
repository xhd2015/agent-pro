## Expected

- After a fast first run exits and a second long-lived session starts, `tty-watch list`
  still shows the live second session (not empty).
- Registry file for the second session still exists after the first `__serve__` grace cleanup.
- List output contains the second session id and command name `sleep`.
- Tolerates table layout (header row and WATCH/ATTACHED columns when implemented).

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
	if resp.SessionID == "" {
		t.Fatal("expected session id from second run")
	}
	if !resp.RegistryExists {
		t.Fatalf("registry should still contain %s after first serve cleanup, ids %v", resp.SessionID, resp.RegistryIDs)
	}
	if strings.TrimSpace(resp.ListOutput) == "" {
		t.Fatalf("list should not be empty while second session is live; registry ids %v", resp.RegistryIDs)
	}
	if !strings.Contains(resp.ListOutput, resp.SessionID) {
		t.Fatalf("list missing second session id %s: %q", resp.SessionID, resp.ListOutput)
	}
	if !strings.Contains(resp.ListOutput, "sleep") {
		t.Fatalf("list missing command sleep: %q", resp.ListOutput)
	}
}
```