## Expected

- JSON parses as Catalog with object `models` (not string array).
- Each model has `id` and `source`; display names match fixtures.
- Dual-listed `grok-4.5` has `source=config.toml` and cache-filled display name.
- No `slug` key appears.

## Errors

- None.

```go
import (
	"encoding/json"
	"strings"
	"testing"

	grokmodels "github.com/xhd2015/agent-pro/agent/grok/models"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertSuccess(t, resp)
	if strings.Contains(string(resp.JSON), `"slug"`) {
		t.Fatalf("unexpected slug key:\n%s", resp.JSON)
	}
	var payload grokmodels.Catalog
	if err := json.Unmarshal(resp.JSON, &payload); err != nil {
		t.Fatalf("unmarshal: %v\n%s", err, string(resp.JSON))
	}
	if payload.Home != req.GrokHome {
		t.Fatalf("home=%q want %q", payload.Home, req.GrokHome)
	}
	if payload.Default != "grok-4.5" {
		t.Fatalf("default=%q", payload.Default)
	}
	if !payload.FromConfig || !payload.FromCache {
		t.Fatalf("from_config=%v from_cache=%v", payload.FromConfig, payload.FromCache)
	}
	want := []grokmodels.Model{
		{ID: "ais-glm-5-2", Source: grokmodels.DefaultConfigFile, DisplayName: "AIS - GLM-5.2"},
		{ID: "grok-4.5", Source: grokmodels.DefaultConfigFile, DisplayName: "Grok 4.5"},
		{ID: "grok-4.6", Source: grokmodels.ModelsCacheFile, DisplayName: "Grok 4.6"},
	}
	if len(payload.Models) != len(want) {
		t.Fatalf("models=%+v want %+v", payload.Models, want)
	}
	for i := range want {
		if payload.Models[i] != want[i] {
			t.Fatalf("models[%d]=%+v want %+v", i, payload.Models[i], want[i])
		}
	}
}
```
