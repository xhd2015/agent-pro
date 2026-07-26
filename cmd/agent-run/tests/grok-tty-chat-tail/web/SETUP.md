# Scenario

**Bug**: web SSE must use page-lifetime `WatchEvents` and deliver late grok-tty events

```
POST grok-tty session -> GET events/stream
SSE must not stop when meta.status becomes finished
```

## Preconditions

- Web server started with `--grok-home` and `--grok-tty-runner-binary` (llm-mock).
- SSE uses `pkgs/agentevents.WatchEvents` file tail.

## Steps

1. Grouping setup sets `req.Mode` for web leaves (`sse` or `sse-finished-append`).

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Runner = "grok-tty"
	return nil
}
```