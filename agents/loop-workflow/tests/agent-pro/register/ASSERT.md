## Expected

- Exit code 0.
- Stdout contains `name: loop-workflow`.
- Stdout is non-empty SKILL.md content (not "unknown skill" error).

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
	if !strings.Contains(resp.Stdout, "name: loop-workflow") {
		t.Fatalf("agent-pro skill show missing frontmatter name:\nstdout:\n%s\nstderr:\n%s", resp.Stdout, resp.Stderr)
	}
	if strings.Contains(resp.Stderr, "unknown skill") {
		t.Fatalf("loop-workflow not registered in knownSkills:\n%s", resp.Stderr)
	}
}
```