## Expected

- JSON `sendable: false`.
- Non-empty `sendable_reason` (codex still working).

```go
import (
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertNoError(t, err)
	if resp.JSONBody == nil { t.Fatal("expected JSON body") }
	sendable, _ := resp.JSONBody["sendable"].(bool)
	if sendable { t.Fatal("expected sendable false for busy screen") }
	reason, _ := resp.JSONBody["sendable_reason"].(string)
	if reason == "" { t.Fatal("expected non-empty sendable_reason") }
}
```
