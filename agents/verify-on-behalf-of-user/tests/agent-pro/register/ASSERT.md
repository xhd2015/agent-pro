---
label: e2e
---

## Expected

- Exit code 0.
- Stdout contains `name: verify-on-behalf-of-user/transcript`.
- Stdout contains `Transcript format rules`.
- Stdout requires inlining full transcript content for direct review.
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
	if !strings.Contains(resp.Stdout, "name: verify-on-behalf-of-user/transcript") {
		t.Fatalf("agent-pro skill show missing frontmatter name:\nstdout:\n%s\nstderr:\n%s", resp.Stdout, resp.Stderr)
	}
	if !strings.Contains(resp.Stdout, "Transcript format rules") {
		t.Fatalf("skill missing transcript section:\nstdout:\n%s\nstderr:\n%s", resp.Stdout, resp.Stderr)
	}
	if !strings.Contains(resp.Stdout, "full") || (!strings.Contains(strings.ToLower(resp.Stdout), "inline") && !strings.Contains(resp.Stdout, "full body") && !strings.Contains(resp.Stdout, "full contents") && !strings.Contains(resp.Stdout, "full content")) {
		t.Fatalf("transcript topic must require inlining full content:\n%s", resp.Stdout)
	}
	if strings.Contains(resp.Stderr, "unknown skill") {
		t.Fatalf("verify-on-behalf-of-user not registered:\n%s", resp.Stderr)
	}
}
```
