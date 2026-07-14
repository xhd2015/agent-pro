## Expected

- Exit code 0.
- Stdout contains `Flash Idea` category.
- Stdout contains `brainstorm` and guideline `agent-pro skill brainstorm show`.

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
	if !strings.Contains(resp.Stdout, "Flash Idea") {
		t.Fatalf("intent-route missing Flash Idea category:\nstdout:\n%s\nstderr:\n%s", resp.Stdout, resp.Stderr)
	}
	if !strings.Contains(resp.Stdout, "brainstorm") {
		t.Fatalf("intent-route missing brainstorm reference:\nstdout:\n%s\nstderr:\n%s", resp.Stdout, resp.Stderr)
	}
	if !strings.Contains(resp.Stdout, "agent-pro skill brainstorm show") {
		t.Fatalf("intent-route missing brainstorm guideline command:\nstdout:\n%s\nstderr:\n%s", resp.Stdout, resp.Stderr)
	}
}
```