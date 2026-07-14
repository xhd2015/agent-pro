## Expected

- Exit code 0.
- Stdout contains `Available skills:` header.
- Stdout lists `followup` with non-empty description text (not name-only line).

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
		if !strings.Contains(line, "followup") {
			continue
		}
		trimmed := strings.TrimSpace(line)
		if len(trimmed) > len("followup") {
			foundWithDescription = true
			break
		}
	}
	if !foundWithDescription {
		t.Fatalf("followup missing from skills list or listed without description:\nstdout:\n%s\nstderr:\n%s", resp.Stdout, resp.Stderr)
	}
}
```