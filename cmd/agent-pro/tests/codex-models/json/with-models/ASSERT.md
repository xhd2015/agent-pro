## Expected

- JSON parses as Catalog with two models (no `gpt-reserve`).
- Each model uses `id` (not `slug`) and `source`.
- `gpt-5.5` has reasoning `[low medium high xhigh]` and default_reasoning `xhigh`.
- `gpt-5.6-sol` includes `ultra` in reasoning.

## Errors

- None.

```go
import (
	"encoding/json"
	"strings"
	"testing"

	codexmodels "github.com/xhd2015/agent-pro/agent/codex/models"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertSuccess(t, resp)
	if strings.Contains(string(resp.JSON), `"slug"`) {
		t.Fatalf("unexpected slug key:\n%s", resp.JSON)
	}
	var payload codexmodels.Catalog
	if err := json.Unmarshal(resp.JSON, &payload); err != nil {
		t.Fatalf("unmarshal: %v\n%s", err, string(resp.JSON))
	}
	if payload.Home != req.CodexHome {
		t.Fatalf("home=%q want %q", payload.Home, req.CodexHome)
	}
	if payload.Default != "gpt-5.5" {
		t.Fatalf("default=%q", payload.Default)
	}
	if !payload.FromConfig || !payload.FromCache {
		t.Fatalf("from_config=%v from_cache=%v", payload.FromConfig, payload.FromCache)
	}
	if len(payload.Models) != 2 {
		t.Fatalf("models=%+v want 2", payload.Models)
	}
	byID := map[string]codexmodels.Model{}
	for _, m := range payload.Models {
		byID[m.ID] = m
	}
	if _, ok := byID["gpt-reserve"]; ok {
		t.Fatalf("hidden id leaked: %+v", payload.Models)
	}
	sol := byID["gpt-5.6-sol"]
	if sol.Source != codexmodels.ModelsCacheFile || sol.DisplayName != "GPT-5.6-Sol" || sol.DefaultReasoning != "medium" {
		t.Fatalf("sol=%+v", sol)
	}
	wantSol := []string{"low", "medium", "high", "xhigh", "max", "ultra"}
	if len(sol.Reasoning) != len(wantSol) {
		t.Fatalf("sol.Reasoning=%v", sol.Reasoning)
	}
	for i := range wantSol {
		if sol.Reasoning[i] != wantSol[i] {
			t.Fatalf("sol.Reasoning=%v", sol.Reasoning)
		}
	}
	g55 := byID["gpt-5.5"]
	want55 := []string{"low", "medium", "high", "xhigh"}
	if g55.Source != codexmodels.ModelsCacheFile || g55.DefaultReasoning != "xhigh" || len(g55.Reasoning) != len(want55) {
		t.Fatalf("gpt-5.5=%+v", g55)
	}
	for i := range want55 {
		if g55.Reasoning[i] != want55[i] {
			t.Fatalf("gpt-5.5.Reasoning=%v", g55.Reasoning)
		}
	}
}
```
