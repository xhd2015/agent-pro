# Scenario

**Feature**: completed turn is idle

```
turn/start + assistant/message + turn/end(completed) -> idle
```

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Events = []EventSpec{
		{Type: "turn/start", Data: `{"turn":1}`},
		{Type: "assistant/message", Data: `{"message":{"content":[{"type":"text","text":"done"}]}}`},
		{Type: "turn/end", Data: `{"turn":1,"reason":{"kind":"completed"}}`},
	}
	req.Live = true
	return nil
}
```
