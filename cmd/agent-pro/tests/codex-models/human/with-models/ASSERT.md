## Expected

- List returns exactly two models (hidden slug excluded); Default is `gpt-5.5`.
- Human output marks `* gpt-5.5` and shows reasoning for both visible models.
- Output does not mention `gpt-reserve`.

## Errors

- None.

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertSuccess(t, resp)
	if resp.Catalog.Default != "gpt-5.5" {
		t.Fatalf("Default=%q", resp.Catalog.Default)
	}
	if len(resp.Catalog.Models) != 2 {
		t.Fatalf("Models=%+v want 2 list-visible", resp.Catalog.Models)
	}
	assertContains(t, resp.Output, "Home: "+req.CodexHome)
	assertContains(t, resp.Output, "Default: gpt-5.5")
	assertContains(t, resp.Output, "* gpt-5.5")
	assertContains(t, resp.Output, "  gpt-5.6-sol")
	assertContains(t, resp.Output, "reasoning=[low medium high xhigh max ultra]")
	assertContains(t, resp.Output, "reasoning=[low medium high xhigh]")
	assertNotContains(t, resp.Output, "gpt-reserve")
	if strings.Contains(resp.Output, "* gpt-5.6-sol") {
		t.Fatalf("non-default marked with *:\n%s", resp.Output)
	}
}
```
