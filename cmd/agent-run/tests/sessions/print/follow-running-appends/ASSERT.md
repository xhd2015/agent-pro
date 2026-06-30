---
label: slow
explanation: subprocess follow loop until meta.status leaves running; sidecar appends after 500ms
---

## Expected Output

```
<contains>
First running event
Following new events
Follow-up appended line
Session finished
</contains>
```

## Expected

- Exit code 0 within subprocess timeout.
- Stdout includes initial event, follow banner, appended event text, and session-finished footer.

## Exit Code

0

```go
import (
	"testing"

	"github.com/xhd2015/doctest/assert"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertSuccess(t, resp)
	assert.Output(t, resp.Stdout, `
<contains>
First running event
Following new events
Follow-up appended line
Session finished
</contains>`)
}
```