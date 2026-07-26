---
label: e2e
---

## Expected

- Exit code 0.
- Stdout contains `name: brainstorm`.
- Stdout contains `CLI output examples` (CLI planning section).
- Stdout is non-empty SKILL.md content (not "unknown skill" error).

## Exit Code

0

```go
import (
	"strings"
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	assertExitCode(t, resp, 0)
	if !strings.Contains(resp.Stdout, "name: brainstorm") {
		t.Fatalf("agent-pro skill show missing frontmatter name:\nstdout:\n%s\nstderr:\n%s", resp.Stdout, resp.Stderr)
	}
	if !strings.Contains(resp.Stdout, "CLI output examples") {
		t.Fatalf("brainstorm skill missing CLI output section:\nstdout:\n%s\nstderr:\n%s", resp.Stdout, resp.Stderr)
	}
	if strings.Contains(resp.Stderr, "unknown skill") {
		t.Fatalf("brainstorm not registered in knownSkills:\n%s", resp.Stderr)
	}
}
```