## Expected
- stdout matches the clean formatted message (sanitize does not alter legitimate content).

## Expected Output

```
---
version: 3
---
feat: keep clean messages intact

No anti-pattern noise
```

```go
import (
	"testing"

	"github.com/xhd2015/doctest/assert"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if resp.Err != nil {
		t.Fatalf("clean JSON should still succeed, got: %v\nstderr:\n%s", resp.Err, resp.Stderr)
	}
	want := ReadAntiPatternWant(t, "clean_json_unchanged")
	AssertStdoutMessage(t, resp.Stdout, want)
	assert.Output(t, resp.Stdout, `---
version: 3
---
feat: keep clean messages intact

No anti-pattern noise
`)
}
```
