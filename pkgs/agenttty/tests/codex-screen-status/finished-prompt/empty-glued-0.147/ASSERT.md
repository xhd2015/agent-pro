## Expected

- `DetectScreenStatus` is `idle` (not `banner`).
- `DetectInputBox` is `empty`.
- `CheckWritable` is ready / `idle`.

## Exit Code

N/A (direct package call)

```go
import "testing"

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	_ = req
	assertScreenIdle(t, resp, err)
	if resp.InputBox != "empty" {
		t.Fatalf("InputBox=%q want empty on 0.147 glued chrome", resp.InputBox)
	}
	if !resp.WritableReady || resp.WritableState != "idle" {
		t.Fatalf("CheckWritable ready=%v state=%q want idle", resp.WritableReady, resp.WritableState)
	}
}
```
