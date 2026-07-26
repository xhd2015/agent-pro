# Scenario

**Feature**: unified AttachRelay across CLI, web WS, and tty-watch

```
CLI attach -> AttachRelay(attach, TTYAttachSink)
web terminal/ws -> AttachRelay(attach, WebSocketAttachSink)
tty-watch attach -> unchanged multi-writer attach semantics
```

## Preconditions

- Fake ptywrap servers accept only `attach_mode=attach` when configured.
- Web terminal websocket must not use snapshot proxy paths.

## Steps

1. Grouping setup sets `req.Area = "attach-relay"`.
2. Leaves start fake ptywrap or live stub/codex sessions as needed.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Area = "attach-relay"
	if req.RegistryTranscript == "" {
		req.RegistryTranscript = "terminal-ready-from-pty"
	}
	if req.WSAttachMode == "" {
		req.WSAttachMode = "attach"
	}
	return nil
}
```
