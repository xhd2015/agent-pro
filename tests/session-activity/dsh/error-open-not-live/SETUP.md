# Scenario

**Feature**: open turn whose process is gone is error

```
turn/start + not live -> error
```

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Events = []EventSpec{{Type: "turn/start", Data: `{"turn":1}`}}
	req.Live = false
	return nil
}
```
