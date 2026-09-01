## Expected

- List succeeds with Default `grok-4.5`.
- Models are the sorted union with correct `source` / `display_name`.
- Human output marks `* grok-4.5` and shows display names.

## Errors

- None.

```go
import (
	"strings"
	"testing"

	grokmodels "github.com/xhd2015/agent-pro/agent/grok/models"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertSuccess(t, resp)
	if resp.Catalog.Default != "grok-4.5" {
		t.Fatalf("Default=%q", resp.Catalog.Default)
	}
	want := []grokmodels.Model{
		{ID: "ais-glm-5-2", Source: grokmodels.DefaultConfigFile, DisplayName: "AIS - GLM-5.2"},
		{ID: "grok-4.5", Source: grokmodels.DefaultConfigFile, DisplayName: "Grok 4.5"},
		{ID: "grok-4.6", Source: grokmodels.ModelsCacheFile, DisplayName: "Grok 4.6"},
	}
	if len(resp.Catalog.Models) != len(want) {
		t.Fatalf("Models=%+v want %+v", resp.Catalog.Models, want)
	}
	for i := range want {
		if resp.Catalog.Models[i] != want[i] {
			t.Fatalf("Models[%d]=%+v want %+v", i, resp.Catalog.Models[i], want[i])
		}
	}
	assertContains(t, resp.Output, "Home: "+req.GrokHome)
	assertContains(t, resp.Output, "Default: grok-4.5")
	assertContains(t, resp.Output, "* grok-4.5  Grok 4.5")
	assertContains(t, resp.Output, "  ais-glm-5-2  AIS - GLM-5.2")
	assertContains(t, resp.Output, "  grok-4.6  Grok 4.6")
	if strings.Contains(resp.Output, "* ais-glm-5-2") || strings.Contains(resp.Output, "* grok-4.6") {
		t.Fatalf("non-default marked with *:\n%s", resp.Output)
	}
}
```
