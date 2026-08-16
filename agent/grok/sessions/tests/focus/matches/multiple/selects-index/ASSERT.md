## Expected

- `--index 1` focuses exactly candidate 1 and prints that window/tab.

## Expected Output

```text
focused: window 3, tab 1
```

```go
import "github.com/xhd2015/doctest/assert"

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	_ = req
	assertNoHarnessErr(t, err)
	assertNoError(t, resp)
	if resp.ListITermCalls != 1 {
		t.Fatalf("ListITermCalls = %d, want 1", resp.ListITermCalls)
	}
	if len(resp.Focused) != 1 || resp.Focused[0] != "w2t1p0" {
		t.Fatalf("Focused = %v, want [w2t1p0]", resp.Focused)
	}
	assert.Output(t, resp.Stdout, `---
version: 3
---
focused: window 3, tab 1
`)
}
```
