## Expected

- `tty status --json` reports `screen_status: idle`.

```go
import (
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertNoError(t, err)
	if resp.JSONBody == nil { t.Fatal("expected JSON") }
	status, _ := resp.JSONBody["screen_status"].(string)
	if status != "idle" { t.Fatalf("screen_status: got %q want idle", status) }
}
```
