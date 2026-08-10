## Expected

- Exactly one publish when URL and token are set.
- Event type/source still `agent.tty.started` / `agent-run`.
- Token is non-empty on the request path:
  - Product `EventBusOpts.Token` is the value from the leaf (`test-token-n3`).
  - Production HTTP path must set `Authorization: Bearer <token>` via
    eventbus.WithToken (covered by eventbus package tests).
- Token must not block publish (PublishCount == 1).

## Side Effects

- One publish.

## Errors

- None.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	t.Helper()
	_ = d
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if req.EventBusToken == "" {
		t.Fatal("leaf fixture requires non-empty EventBusToken")
	}
	if resp.PublishCount != 1 {
		t.Fatalf("PublishCount: got %d, want 1 (token must not block publish)", resp.PublishCount)
	}
	got, ok := req.Capture.Last()
	if !ok {
		t.Fatal("missing captured publish")
	}
	if got.Type != wireTypeAgentTTYStarted || got.Source != wireSourceAgentRun {
		t.Fatalf("type/source: got %q/%q", got.Type, got.Source)
	}
}
```
