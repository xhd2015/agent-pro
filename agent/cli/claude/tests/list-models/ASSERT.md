## Expected
- `Response.Models` is nil (claude has no model-listing command; `ListModels`
  returns `nil, nil`).
- No error occurred.

## Side Effects
- None.

## Exit Code
- Not applicable (in-process agent call, not a CLI invocation).

```go
import (
	"testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("ListModels failed: %v", err)
	}
	if resp.Models != nil {
		t.Fatalf("expected nil model list (unsupported), got: %v", resp.Models)
	}
}
```
