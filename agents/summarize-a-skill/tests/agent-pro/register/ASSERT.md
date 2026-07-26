---
label: e2e
---

## Expected

- Exit code 0.
- Stdout contains `name: summarize-a-skill`.
- Stdout contains `Output path resolution`.
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
	if !strings.Contains(resp.Stdout, "name: summarize-a-skill") {
		t.Fatalf("agent-pro skill show missing frontmatter name:\nstdout:\n%s\nstderr:\n%s", resp.Stdout, resp.Stderr)
	}
	if !strings.Contains(resp.Stdout, "Output path resolution") {
		t.Fatalf("summarize-a-skill skill missing path resolution section:\nstdout:\n%s\nstderr:\n%s", resp.Stdout, resp.Stderr)
	}
	if strings.Contains(resp.Stderr, "unknown skill") {
		t.Fatalf("summarize-a-skill not registered in knownSkills:\n%s", resp.Stderr)
	}
}
```
