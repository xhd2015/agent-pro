## Expected
- Each skill tool call prints a `SKILL` header with its selected skill name.
- Bare `SKILL` headers without `confluence-fetch` or `git-fetch` are insufficient.

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertContains(t, resp.Output, "SKILL")
	assertContains(t, resp.Output, "confluence-fetch")
	assertContains(t, resp.Output, "git-fetch")

	skillHeaders := strings.Count(resp.Output, "SKILL")
	if skillHeaders != 2 {
		t.Fatalf("expected 2 SKILL blocks, got %d:\n%s", skillHeaders, resp.Output)
	}
}
```