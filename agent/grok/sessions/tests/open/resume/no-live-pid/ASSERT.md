## Expected

- No live host → resume in new window; iTerm list skipped.

## Expected Output

```text
opened: new window; resuming <id>
```

```go
import "github.com/xhd2015/doctest/assert"

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertNoHarnessErr(t, err)
	assertNoError(t, resp)
	if resp.ListITermCalls != 0 {
		t.Fatalf("ListITermCalls = %d, want 0", resp.ListITermCalls)
	}
	assertResumeOpened(t, req, resp)
	assert.Output(t, resp.Stdout, `---
version: 3
---
opened: new window; resuming `+req.SessionID+`
`)
}
```
