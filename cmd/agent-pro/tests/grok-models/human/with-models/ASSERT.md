## Expected

- List succeeds with Default `grok-4.5`.
- Models are the sorted union: `ais-glm-5-2`, `grok-4.5`, `grok-4.6`.
- Human output contains `Home:` and `Default: grok-4.5`.
- Default line is marked `* grok-4.5`; others use two-space indent.

## Errors

- None.

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertSuccess(t, resp)
	if resp.Catalog.Default != "grok-4.5" {
		t.Fatalf("Default=%q", resp.Catalog.Default)
	}
	want := []string{"ais-glm-5-2", "grok-4.5", "grok-4.6"}
	if len(resp.Catalog.Models) != len(want) {
		t.Fatalf("Models=%v want %v", resp.Catalog.Models, want)
	}
	for i := range want {
		if resp.Catalog.Models[i] != want[i] {
			t.Fatalf("Models=%v want %v", resp.Catalog.Models, want)
		}
	}
	assertContains(t, resp.Output, "Home: "+req.GrokHome)
	assertContains(t, resp.Output, "Default: grok-4.5")
	assertContains(t, resp.Output, "* grok-4.5")
	assertContains(t, resp.Output, "  ais-glm-5-2")
	assertContains(t, resp.Output, "  grok-4.6")
	if strings.Contains(resp.Output, "* ais-glm-5-2") || strings.Contains(resp.Output, "* grok-4.6") {
		t.Fatalf("non-default marked with *:\n%s", resp.Output)
	}
}
```
