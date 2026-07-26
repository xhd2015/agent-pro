# Scenario

**Feature**: --keep-tty keeps registry and tty.json alive

```
run --keep-tty -> registry + tty.json alive=true after exit
```

## Steps

1. Configure leaf-specific `req` fields.
2. `Run` executes the scenario.

```go
import (
	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Action = "keep-tty-persists"
	return nil
}
```
