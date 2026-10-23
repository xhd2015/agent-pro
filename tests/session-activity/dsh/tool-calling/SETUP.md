# Scenario

**Feature**: unmatched tool/call is tool_calling

```
turn/start + tool/call(bash) + live -> tool_calling detail=bash
```

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Events = []EventSpec{
		{Type: "turn/start", Data: `{"turn":1}`},
		{Type: "tool/call", Data: `{"name":"bash"}`},
	}
	req.Live = true
	return nil
}
```
