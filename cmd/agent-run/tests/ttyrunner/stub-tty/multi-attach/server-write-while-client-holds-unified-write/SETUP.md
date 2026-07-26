# Scenario

**Feature**: server send works while client holds unified write

```
writer attached + tty send -> both succeed (two write planes)
```

## Steps

1. Configure leaf-specific `req` fields.
2. `Run` executes the scenario.

```go
import (
	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Action = "server-write-while-client-holds-unified-write"
	return nil
}
```
