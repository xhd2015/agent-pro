---
label: e2e
---

## Expected

- Exit code 0.
- Stdout is valid JSON with pid, listen_addr, tty_type, session_id, created_at, tcp_reachable keys.

```go
import (
	"testing"
	"strings"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if resp.JSONBody == nil {
		t.Fatal("expected JSON body, got nil")
	}

	keys := map[string]bool{"pid": true, "port": true, "tty_type": true, "session_id": true, "start_time": true}
	for k := range keys {
		if _, ok := resp.JSONBody[k]; !ok {
			t.Fatalf("JSON missing key %q in:\n%s", k, resp.Stdout)
		}
	}

	ttyType, _ := resp.JSONBody["tty_type"].(string)
	if !strings.EqualFold(ttyType, "grok-tty") {
		t.Fatalf("expected tty_type grok-tty, got %q", ttyType)
	}
}
```
