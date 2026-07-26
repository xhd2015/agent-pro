## Expected
- Response.Models is non-empty (has at least one model).
- Response.Models contains `"grok-composer-2.5-fast"`.
- No error occurred.

## Side Effects
- None.

## Exit Code
- Not applicable (in-process agent call, not a CLI invocation).

```go
import (
	"slices"
	"testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("ListModels failed: %v", err)
	}
	if len(resp.Models) == 0 {
		t.Fatal("expected non-empty model list")
	}
	if !slices.Contains(resp.Models, "grok-composer-2.5-fast") {
		t.Fatalf("expected model list to contain 'grok-composer-2.5-fast', got: %v", resp.Models)
	}
}
```
