---
label: e2e
---

## Expected

- Exit code 0.
- `provider.multiprov.models` has exactly two keys: `m1` and `m2`.
- Each model value is `{"name": "<model-id>"}` (name equals the model id).

## Side Effects

- Global config file created with both model entries.

## Errors

- None.

## Exit Code

- 0.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	assertSuccessCommon(t, req, resp, "multiprov")
	entry, _, e := readProviderEntry(resp.ConfigPath, "multiprov")
	if e != nil {
		t.Fatal(e)
	}
	models, _ := entry["models"].(map[string]interface{})
	if len(models) != 2 {
		t.Fatalf("models has %d entries, want 2: %v", len(models), models)
	}
	for _, mid := range []string{"m1", "m2"} {
		m, ok := models[mid]
		if !ok {
			t.Fatalf("models.%s missing: %v", mid, models)
		}
		mm, _ := m.(map[string]interface{})
		if mm["name"] != mid {
			t.Fatalf("models.%s.name = %v, want %s", mid, mm["name"], mid)
		}
	}
}
```
