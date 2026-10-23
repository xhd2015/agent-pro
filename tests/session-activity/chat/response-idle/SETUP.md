# Scenario

**Feature**: last response is idle even if the TUI process is still live

```
response + live -> idle
```

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.LastKind = "response"
	req.Live = true
	return nil
}
```
