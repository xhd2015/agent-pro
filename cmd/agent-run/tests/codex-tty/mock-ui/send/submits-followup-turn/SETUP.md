# Scenario

**Bug**: `agent-run send` injects `text\r` in one write; Codex TUI only types into
the composer and does not start a turn.

```
open idle → send "follow-up-two" → SECOND_MOCK_REPLY (not › follow-up-two draft-only)
```

## Steps

1. `Mode=mock-ui-send`, submit follow-up-two, expect mock reply.

```go
import (
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Mode = "mock-ui-send"
	req.SessionID = "mock-ui-send-submit"
	req.MockUISendText = "follow-up-two"
	req.MockUIExpectReply = "SECOND_MOCK_REPLY"
	req.MockUINoSubmit = false
	return nil
}
```
