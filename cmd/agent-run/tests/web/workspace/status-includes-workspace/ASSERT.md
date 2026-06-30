## Expected

- HTTP **200**.
- JSON body includes non-empty `workspace` and `home` keys.

```go
import (
	"encoding/json"
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.HTTPStatus != 200 {
		t.Fatalf("expected HTTP 200, got %d body=%q", resp.HTTPStatus, resp.HTTPBody)
	}
	var parsed map[string]any
	if err := json.Unmarshal([]byte(resp.HTTPBody), &parsed); err != nil {
		t.Fatalf("parse status JSON: %v body=%q", err, resp.HTTPBody)
	}
	home, _ := parsed["home"].(string)
	workspace, _ := parsed["workspace"].(string)
	if strings.TrimSpace(home) == "" {
		t.Fatalf("status.home empty: %q", resp.HTTPBody)
	}
	if strings.TrimSpace(workspace) == "" {
		t.Fatalf("status.workspace empty: %q", resp.HTTPBody)
	}
}
```