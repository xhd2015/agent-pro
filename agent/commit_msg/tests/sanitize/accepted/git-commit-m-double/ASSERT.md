## Expected
- stdout is title plus description paragraphs joined by blank lines (from 2nd+ `-m`).

## Expected Output

```
---
version: 3
---
feat: split multi -m into title and body

First paragraph of description\.

Second paragraph of description\.
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
		t.Fatalf("gen-commit-msg should split multi -m, got: %v\nstderr:\n%s", resp.Err, resp.Stderr)
	}
	want := ReadAntiPatternWant(t, "git_commit_m_double")
	AssertStdoutMessage(t, resp.Stdout, want)
	assert.Output(t, resp.Stdout, `---
version: 3
---
feat: split multi -m into title and body

First paragraph of description\.

Second paragraph of description\.
`)
	if strings.Contains(resp.Stdout, "git commit") {
		t.Fatalf("git commit wrapper leaked into stdout:\n%s", resp.Stdout)
	}
}
```
