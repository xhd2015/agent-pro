---
label: e2e
---

## Expected

- Exit code 0.
- Stdout contains `Available skills:` header.
- Stdout lists `run-the-loop` with non-empty description text (not name-only line).

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
	if !strings.Contains(resp.Stdout, "Available skills:") {
		t.Fatalf("skills list missing header:\nstdout:\n%s\nstderr:\n%s", resp.Stdout, resp.Stderr)
	}
	var foundWithDescription bool
	for _, line := range strings.Split(resp.Stdout, "\n") {
		if !strings.Contains(line, "run-the-loop") {
			continue
		}
		trimmed := strings.TrimSpace(line)
		if len(trimmed) > len("run-the-loop") {
			foundWithDescription = true
			break
		}
	}
	if !foundWithDescription {
		t.Fatalf("run-the-loop missing from skills list or listed without description:\nstdout:\n%s\nstderr:\n%s", resp.Stdout, resp.Stderr)
	}
}
```