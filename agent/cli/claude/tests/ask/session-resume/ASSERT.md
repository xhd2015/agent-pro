---
label: slow, heavy
explanation: runs the real claude binary headless twice; LLM latency + cost
---

## Expected
- `Response.Answer` is non-empty (the resumed turn produced a response).
- `Response.SessionID` is non-empty (session persisted across turns).
- No error occurred.

## Side Effects
- Spawns the real `claude` binary twice; consumes LLM tokens.

## Exit Code
- Not applicable (in-process agent call, not a CLI invocation).

```go
import (
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("run failed: %v", err)
	}
	if resp.SessionID == "" {
		t.Fatal("expected non-empty SessionID after resumed Ask()")
	}
	if resp.Answer == "" {
		t.Fatal("expected non-empty answer from resumed query")
	}
}
```
