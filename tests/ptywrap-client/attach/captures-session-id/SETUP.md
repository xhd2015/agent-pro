# Scenario

**Feature**: Attach captures session_id from server WS message

```
# session_id handshake
WS connect -> server session_id JSON -> AttachResult.SessionID
```

```go
import (
	"testing"
	"time"

	"github.com/xhd2015/dot-pkgs/go-pkgs/shell/ptywrap"
)

func Setup(t *testing.T, req *Request) error {
	req.Phase = "attach-captures-id"
	req.Sessions = []ptywrap.SessionInfo{{
		ID:        "session-42",
		Name:      "capture-test",
		Status:    "running",
		CreatedAt: time.Now(),
	}}
	req.ConnectOpts.SessionID = "session-42"
	return nil
}
```