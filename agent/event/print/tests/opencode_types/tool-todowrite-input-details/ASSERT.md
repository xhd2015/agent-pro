## Expected
- Output contains the `TODO` header.
- Output includes each todo item content.
- Output includes status markers or status text so the plan is not just a bare
  tool heading.

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(resp.Output, "TODO") && !strings.Contains(resp.Output, "PLAN") {
		t.Fatalf("expected TODO/PLAN header in:\n%s", resp.Output)
	}
	assertContains(t, resp.Output, "Inspect pricing BFF routes")
	assertContains(t, resp.Output, "Update credit.spl.bff glossary")
	assertContains(t, resp.Output, "pending")
}
```
