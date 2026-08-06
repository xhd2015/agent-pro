## Expected

- No error.
- Output is a JSON array of objects each with `name` (and preferably `size`).
- All three basenames appear.
- No ANSI escapes.

## Errors

- None.

```go
import (
	"encoding/json"
	"testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertNoHarnessErr(t, err)
	assertNoError(t, resp)
	if resp.Output == "" {
		t.Fatal("json output empty")
	}
	assertNoANSI(t, resp.Output)

	var arr []map[string]any
	if err := json.Unmarshal([]byte(resp.Output), &arr); err != nil {
		t.Fatalf("json.Unmarshal: %v\nraw=%s", err, resp.Output)
	}
	if len(arr) < 3 {
		t.Fatalf("json array len = %d, want >= 3", len(arr))
	}
	got := map[string]bool{}
	for _, obj := range arr {
		name, _ := obj["name"].(string)
		if name != "" {
			got[name] = true
		}
	}
	for _, want := range []string{"summary.json", "updates.jsonl", "signals.json"} {
		if !got[want] {
			t.Fatalf("json missing name %q in %v", want, got)
		}
	}
}
```
