## Expected

- `DetectScreenStatus` is `idle` (not `banner`).
- `DetectInputBox` is `empty` (` medium · ` glue).
- `CheckWritable` is ready / `idle`.

## Exit Code

N/A (direct package call)

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	_ = req
	assertScreenIdle(t, resp, err)
	if !strings.Contains(resp.Text, " medium · ") {
		t.Fatalf("host fixture must contain footer glue, got %q", resp.Text)
	}
	if resp.InputBox != "empty" {
		t.Fatalf("InputBox=%q want empty on host medium · chrome", resp.InputBox)
	}
	if !resp.WritableReady || resp.WritableState != "idle" {
		t.Fatalf("CheckWritable ready=%v state=%q want idle", resp.WritableReady, resp.WritableState)
	}
}
```
