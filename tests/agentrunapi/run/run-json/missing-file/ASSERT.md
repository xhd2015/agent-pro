## Expected

- API error mentions the result file / JSON.

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertNoError(t, err)
	assertAPIError(t, resp)
	lower := strings.ToLower(resp.ErrString)
	if !strings.Contains(lower, "result file") && !strings.Contains(lower, "json") {
		t.Fatalf("error should mention result file/json, got %q", resp.ErrString)
	}
	if resp.LaunchResultFile != "" {
		assertContains(t, resp.ErrString, resp.LaunchResultFile)
	}
}
```
