## Expected

- `run --session-id` succeeds after stale seed (exit 0 from list probe).
- Registry exists for custom id with reachable listen address.
- `list` output includes `test-with-grok`.

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
		t.Fatalf("expected reused custom id test-with-grok, got %q", resp.SessionID)
	}
	if !resp.RegistryExists {
		t.Fatalf("registry should exist for reused id %s", resp.SessionID)
	}
	if !resp.SessionRunning {
		t.Fatalf("reused session %s should be reachable", resp.SessionID)
	}
	if !strings.Contains(resp.ListOutput, "test-with-grok") {
		t.Fatalf("list missing reused custom id: %q", resp.ListOutput)
	}
}
```