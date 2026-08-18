## Expected
- stdout is the cleaned title only (no `**Title (N chars):**` meta, no outer ticks).

## Expected Output

```
---
version: 3
---
feat: improve commit message parsing
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
		t.Fatalf("gen-commit-msg should succeed after stripping md meta, got: %v\nstderr:\n%s", resp.Err, resp.Stderr)
	}
	want := ReadAntiPatternWant(t, "md_title_char_annotation")
	AssertStdoutMessage(t, resp.Stdout, want)
	assert.Output(t, resp.Stdout, `---
version: 3
---
feat: improve commit message parsing
`)
	if strings.Contains(resp.Stdout, "Title") || strings.Contains(resp.Stdout, "chars") {
		t.Fatalf("md meta leaked into stdout:\n%s", resp.Stdout)
	}
}
```
