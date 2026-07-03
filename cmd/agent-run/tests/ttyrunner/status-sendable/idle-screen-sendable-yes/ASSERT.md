## Expected

- JSON `sendable: true`.
- `sendable_reason` empty when sendable.

```go
import (
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertNoError(t, err)
	if resp.ExitCode != 0 { t.Fatalf("exit %d stderr:\n%s", resp.ExitCode, resp.Stderr) }
	if resp.JSONBody == nil { t.Fatal("expected JSON body") }
	sendable, ok := resp.JSONBody["sendable"].(bool)
	if !ok || !sendable { t.Fatalf("sendable: got %v want true", resp.JSONBody["sendable"]) }
}
```
