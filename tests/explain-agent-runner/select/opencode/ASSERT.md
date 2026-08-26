---
label: e2e
---

## Expected

- Exit 0.
- Stdout contains `[MOCK OK]`.
- Persisted `agent_runner` is `opencode`.
- Meta contains session id `fake-sess-1`.

```go
import (
	"encoding/json"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	_ = req
	if err != nil {
		t.Fatal(err)
	}
	assertExitCode(t, resp, 0)
	assertContains(t, resp.Stdout, "[MOCK OK]")
	if resp.AgentRunner != "opencode" {
		t.Fatalf("AgentRunner = %q, want opencode\nstderr:\n%s", resp.AgentRunner, resp.Stderr)
	}
	var fields map[string]string
	if err := json.Unmarshal(resp.SessionMeta, &fields); err != nil {
		t.Fatalf("meta unmarshal: %v (%s)", err, string(resp.SessionMeta))
	}
	if fields["session_id"] != "fake-sess-1" {
		t.Fatalf("session_id = %q, want fake-sess-1 (meta=%s)", fields["session_id"], string(resp.SessionMeta))
	}
}
```