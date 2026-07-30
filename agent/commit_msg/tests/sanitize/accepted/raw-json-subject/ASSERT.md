## Expected
- stdout is formatted title + blank line + description (not the raw JSON object).

## Expected Output

```
---
version: 3
---
feat: add dir upload with retry

Stream tar upload and resume failed chunks
```

```go
import (
	"strings"
	"testing"

	"github.com/xhd2015/doctest/assert"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if resp.Err != nil {
		t.Fatalf("gen-commit-msg should format raw JSON, got: %v\nstderr:\n%s", resp.Err, resp.Stderr)
	}
	want := ReadAntiPatternWant(t, "json_raw_subject")
	AssertStdoutMessage(t, resp.Stdout, want)
	assert.Output(t, resp.Stdout, `---
version: 3
---
feat: add dir upload with retry

Stream tar upload and resume failed chunks
`)
	if strings.Contains(resp.Stdout, `"title"`) || strings.HasPrefix(strings.TrimSpace(resp.Stdout), "{") {
		t.Fatalf("raw JSON leaked as commit message:\n%s", resp.Stdout)
	}
}
```
