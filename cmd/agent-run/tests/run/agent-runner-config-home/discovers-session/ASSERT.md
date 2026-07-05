## Expected

- Exit code 0.
- Stderr contains `grok-tty: grok session b2222222-2222-4222-8222-222222222222` and
  `grok-tty: grok updates` with the seeded `updates.jsonl` path.
- Stdout contains streamed marker `CONFIG_HOME_STREAM_MARKER`.

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
	assertSuccess(t, resp)
	assertStderrGrokSession(t, resp.Stderr, configHomeUUID, req.GrokUpdatesPath)
	if !strings.Contains(resp.Stdout, "CONFIG_HOME_STREAM_MARKER") {
		t.Fatalf("stdout missing stream marker; stdout:\n%s", resp.Stdout)
	}
}
```