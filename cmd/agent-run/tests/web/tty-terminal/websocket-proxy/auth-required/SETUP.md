# Scenario

**Feature**: terminal websocket attach requires agent-run API auth

```
browser WS attach without Bearer -> 401/403, no upstream attach
```

## Preconditions

- Web server was started with `--token test`.

## Steps

1. Clear websocket auth header.
2. Try to attach to terminal websocket.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.WSAuth = ""
	req.WSInput = ""
	return nil
}
```
