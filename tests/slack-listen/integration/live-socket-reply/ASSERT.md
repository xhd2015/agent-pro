---
label: integration, slow
explanation: live Socket Mode connect probe with repo slack-config.json
---

## Expected

- Daemon starts without immediate token/config error.
- Output contains `Using config from: ` + absolute repo config path.
- Process stops on SIGTERM without panic.

## Side Effects

- Opens live Socket Mode connection (no guaranteed inbound message in CI).

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
	if strings.Contains(combined, "bot token required") || strings.Contains(combined, "app token required") {
		t.Fatalf("live config should supply tokens, got:\n%s", combined)
	}
}
```