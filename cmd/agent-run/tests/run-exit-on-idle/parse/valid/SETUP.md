# Scenario

**Feature**: valid idle-flag parse does not start a TTY

```
ParseRunIdle(false, "2s") -> enabled=false, no error
```

## Steps

1. Grouping documents the valid family.
2. Leaf sets a parseable timeout without `--exit-on-idle`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.Op = opParse
	return nil
}
```
