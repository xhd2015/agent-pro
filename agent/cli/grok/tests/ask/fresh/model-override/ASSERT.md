## Expected
- Response.Answer is non-empty (the model responded).
- No error occurred.

## Side Effects
- None.

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
	if resp.Answer == "" {
		t.Fatal("expected non-empty answer from model override query")
	}
}
```
