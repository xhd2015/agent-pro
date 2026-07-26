## Expected
- stdout keeps inner code spans (`` `--open` `` and `` `git` ``) while remaining a clean message.

## Expected Output

```
---
version: 2
---
feat: add `--open` flag

Pass path to `git` when needed
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
		t.Fatalf("inner backticks case should succeed, got: %v\nstderr:\n%s", resp.Err, resp.Stderr)
	}
	want := ReadAntiPatternWant(t, "legitimate_inner_backticks")
	AssertStdoutMessage(t, resp.Stdout, want)
	assert.Output(t, resp.Stdout, "---\nversion: 2\n---\nfeat: add `--open` flag\n\nPass path to `git` when needed\n")
	if !strings.Contains(resp.Stdout, "`--open`") {
		t.Fatalf("legitimate inner backticks around --open were stripped:\n%s", resp.Stdout)
	}
	if !strings.Contains(resp.Stdout, "`git`") {
		t.Fatalf("legitimate inner backticks around git were stripped:\n%s", resp.Stdout)
	}
}
```
