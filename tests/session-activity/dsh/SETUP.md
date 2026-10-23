# Scenario

**Feature**: DSH durable events

```
DSH events + live -> activity
```

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Source = "dsh"
	return nil
}
```
