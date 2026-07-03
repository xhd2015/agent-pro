## Expected

- Exit code 0.
- Stdout contains `name: debug-with-user` in YAML frontmatter.
- Stdout mentions human-assisted / macOS debugging workflow keywords.

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
	if !strings.Contains(resp.Stdout, "name: debug-with-user") {
		t.Fatalf("skill show missing frontmatter name:\n%s", resp.Stdout)
	}
	lower := strings.ToLower(resp.Stdout)
	if !strings.Contains(lower, "debug-with-user") && !strings.Contains(lower, "human") {
		t.Fatalf("skill show missing workflow description:\n%s", resp.Stdout)
	}
}
```
