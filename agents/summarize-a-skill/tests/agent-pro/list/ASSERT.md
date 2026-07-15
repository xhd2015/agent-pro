## Expected

- Exit code 0.
- Stdout contains `Available skills:` header.
- Stdout lists `summarize-a-skill` with non-empty description text (not name-only line).

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
	if !strings.Contains(resp.Stdout, "Available skills:") {
		t.Fatalf("skills list missing header:\nstdout:\n%s\nstderr:\n%s", resp.Stdout, resp.Stderr)
	}
	var foundWithDescription bool
	for _, line := range strings.Split(resp.Stdout, "\n") {
		if !strings.Contains(line, "summarize-a-skill") {
			continue
		}
		trimmed := strings.TrimSpace(line)
		if len(trimmed) > len("summarize-a-skill") {
			foundWithDescription = true
			break
		}
	}
	if !foundWithDescription {
		t.Fatalf("summarize-a-skill missing from skills list or listed without description:\nstdout:\n%s\nstderr:\n%s", resp.Stdout, resp.Stderr)
	}
}
```
