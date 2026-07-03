## Expected

- `sendable: false`.
- `sendable_reason` mentions unreachable or no terminal output.

```go
import (
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertNoError(t, err)
	if resp.JSONBody == nil { t.Fatal("expected JSON body") }
	sendable, _ := resp.JSONBody["sendable"].(bool)
	if sendable { t.Fatal("expected sendable false when unreachable") }
	reason, _ := resp.JSONBody["sendable_reason"].(string)
	if reason == "" { t.Fatal("expected sendable_reason for unreachable") }
}
```
