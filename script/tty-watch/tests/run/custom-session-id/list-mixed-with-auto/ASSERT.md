## Expected

- Exit code 0 from `list`.
- `list` output contains custom id `test-with-grok`.
- `list` output contains an auto-reserved `session-N` id.
- Registry dir has at least two entries including both id styles.

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
	if !resp.RegistryExists {
		t.Fatal("expected both custom and auto registry entries")
	}
	if !resp.SessionRunning {
		t.Fatal("expected both sessions reachable")
	}
	out := resp.ListOutput
	if !strings.Contains(out, "test-with-grok") {
		t.Fatalf("list missing custom id test-with-grok: %q", out)
	}
	hasAuto := false
	for _, id := range resp.RegistryIDs {
		if strings.HasPrefix(id, "session-") {
			hasAuto = true
			if !strings.Contains(out, id) {
				t.Fatalf("list missing auto session id %s: %q", id, out)
			}
		}
	}
	if !hasAuto {
		t.Fatalf("registry should include auto session-N id, got %v", resp.RegistryIDs)
	}
}
```