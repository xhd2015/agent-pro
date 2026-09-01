## Expected

- JSON Catalog has empty `models` and `from_config`/`from_cache` false.
- `home` matches the synthetic path.

## Errors

- None.

```go
import (
	"encoding/json"
	"testing"

	codexmodels "github.com/xhd2015/agent-pro/agent/codex/models"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertSuccess(t, resp)
	var payload codexmodels.Catalog
	if err := json.Unmarshal(resp.JSON, &payload); err != nil {
		t.Fatalf("unmarshal: %v\n%s", err, string(resp.JSON))
	}
	if payload.Home != req.CodexHome {
		t.Fatalf("home=%q want %q", payload.Home, req.CodexHome)
	}
	if payload.Default != "" || len(payload.Models) != 0 {
		t.Fatalf("payload=%+v", payload)
	}
	if payload.FromConfig || payload.FromCache {
		t.Fatalf("from flags should be false: %+v", payload)
	}
}
```
