## Expected

- JSON parses as Catalog with `home`, `default`, `models`, `from_config`, `from_cache`.
- `default` is `grok-4.5`; models are the sorted union of three ids.
- `from_config` and `from_cache` are true.

## Errors

- None.

```go
import (
	"encoding/json"
	"testing"

	grokmodels "github.com/xhd2015/agent-pro/agent/grok/models"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertSuccess(t, resp)
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
	want := []string{"ais-glm-5-2", "grok-4.5", "grok-4.6"}
	if len(payload.Models) != len(want) {
		t.Fatalf("models=%v want %v", payload.Models, want)
	}
	for i := range want {
		if payload.Models[i] != want[i] {
			t.Fatalf("models=%v want %v", payload.Models, want)
		}
	}
}
```
