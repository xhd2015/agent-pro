# Scenario

**Feature**: `send --no-submit` injects without trailing Enter.

```
open idle → send --no-submit "draft-only-text" → still in ›, no mock reply
```

## Steps

1. `Mode=mock-ui-send` with `MockUINoSubmit=true`.

```go
import (
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Mode = "mock-ui-send"
	req.SessionID = "mock-ui-send-nosubmit"
	req.MockUISendText = "draft-only-text"
	req.MockUINoSubmit = true
	return nil
}
```
