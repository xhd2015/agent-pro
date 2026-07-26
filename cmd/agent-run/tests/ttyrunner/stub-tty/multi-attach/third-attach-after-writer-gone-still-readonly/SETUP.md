# Scenario

**Feature**: third attach after writer gone is still read-only

```
writer gone -> third interactive attach -> permanent observer
```

## Steps

1. Configure leaf-specific `req` fields.
2. `Run` executes the scenario.

```go
import (
	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Action = "third-attach-after-writer-gone-still-readonly"
	return nil
}
```
