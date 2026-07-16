## Expected

- Stdout first line is `session-id: my-job`.
- Registry file `my-job.json` exists.

## Exit Code

0

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
	if resp.ExitCode != 0 {
		t.Fatalf("expected exit 0, got %d stderr=%q", resp.ExitCode, resp.Stderr)
	}
	assert.Output(t, strings.TrimSpace(resp.Stdout), `---
version: 3
---
session-id: my-job`)
	if resp.SessionID != "my-job" {
		t.Fatalf("expected session id my-job, got %q", resp.SessionID)
	}
	assertRegistryHasSession(t, req.TTYWatchHome, "my-job")
}
```