## Expected

- Error mentions session not found or expired.

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err == nil && resp.Err == nil { t.Fatal("expected not found error") }
	combined := ""
	if err != nil { combined += err.Error() }
	if resp.Err != nil { combined += resp.Err.Error() }
	lower := strings.ToLower(combined)
	if !strings.Contains(lower, "not found") && !strings.Contains(lower, "expired") {
		t.Fatalf("expected not found/expired error, got: %s", combined)
	}
}
```
