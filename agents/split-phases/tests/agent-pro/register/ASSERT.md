## Expected

- Exit code 0.
- Stdout contains `name: split-phases`.
- Stdout contains `independently implementable` instruction.
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
	if !strings.Contains(resp.Stdout, "name: split-phases") {
		t.Fatalf("agent-pro skill show missing frontmatter name:\nstdout:\n%s\nstderr:\n%s", resp.Stdout, resp.Stderr)
	}
	if !strings.Contains(resp.Stdout, "independently implementable") {
		t.Fatalf("split-phases skill missing independently implementable instruction:\nstdout:\n%s\nstderr:\n%s", resp.Stdout, resp.Stderr)
	}
	if strings.Contains(resp.Stderr, "unknown skill") {
		t.Fatalf("split-phases not registered in knownSkills:\n%s", resp.Stderr)
	}
}
```
