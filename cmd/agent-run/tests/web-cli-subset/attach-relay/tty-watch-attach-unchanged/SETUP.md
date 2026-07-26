# Scenario

**Feature**: tty-watch attach semantics unchanged after AttachRelay extraction

```
tty-watch run --detach cat -> tty-watch attach + stdin -> marker visible
```

## Steps

1. Run detached tty-watch session.
2. Attach with stdin marker via tty-watch attach mode.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Mode = "tty-watch-attach"
	return nil
}
```
