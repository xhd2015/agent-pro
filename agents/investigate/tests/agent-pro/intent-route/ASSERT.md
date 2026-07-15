## Expected

- Exit code 0.
- Stdout contains `Investigation` category.
- Stdout contains `investigate` and guideline `agent-pro skill investigate --show`.

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
	assertExitCode(t, resp, 0)
	if !strings.Contains(resp.Stdout, "Investigation") {
		t.Fatalf("intent-route missing Investigation category:\nstdout:\n%s\nstderr:\n%s", resp.Stdout, resp.Stderr)
	}
	if !strings.Contains(resp.Stdout, "investigate") {
		t.Fatalf("intent-route missing investigate reference:\nstdout:\n%s\nstderr:\n%s", resp.Stdout, resp.Stderr)
	}
	if !strings.Contains(resp.Stdout, "agent-pro skill investigate --show") {
		t.Fatalf("intent-route missing investigate guideline command:\nstdout:\n%s\nstderr:\n%s", resp.Stdout, resp.Stderr)
	}
}
```