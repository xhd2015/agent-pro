# Scenario

**Feature**: last thinking message while live is thinking

```
thinking + live -> thinking
```

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.LastKind = "thinking"
	req.Live = true
	return nil
}
```
