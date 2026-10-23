# Scenario

**Feature**: open turn with no assistant text is thinking

```
turn/start + live -> thinking
```

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Events = []EventSpec{{Type: "turn/start", Data: `{"turn":1}`}}
	req.Live = true
	return nil
}
```
