## Expected

- Stdout first line is `session-id: my-job`.
- Registry file `my-job.json` exists.

```go
import (
	"strings"
	"testing"

	"github.com/xhd2015/doctest/assert"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	assert.Output(t, strings.TrimSpace(resp.Stdout), `---
version: 2
---
session-id: my-job`)
	if resp.SessionID != "my-job" {
		t.Fatalf("expected session id my-job, got %q", resp.SessionID)
	}
	assertRegistryHasSession(t, req.TTYWatchHome, "my-job")
}
```