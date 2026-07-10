---
label: unit
explanation: daemon startup log with explicit --config
---

## Expected

- Startup output contains `Using config from: ` followed by absolute path to materialized config.
- Path matches `req.ConfigPath`.

## Exit Code

0

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	combined := resp.Stdout + resp.Stderr
	want := "Using config from: " + req.ConfigPath
	if !strings.Contains(combined, want) {
		t.Fatalf("output missing %q\nstdout:\n%s\nstderr:\n%s", want, resp.Stdout, resp.Stderr)
	}
}
```