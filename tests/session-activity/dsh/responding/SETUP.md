# Scenario

**Feature**: assistant text before turn/end is responding

```
turn/start + assistant/message(text) + live -> responding
```

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Events = []EventSpec{
		{Type: "turn/start", Data: `{"turn":1}`},
		{Type: "assistant/message", Data: `{"message":{"content":[{"type":"text","text":"hello"}]}}`},
	}
	req.Live = true
	return nil
}
```
