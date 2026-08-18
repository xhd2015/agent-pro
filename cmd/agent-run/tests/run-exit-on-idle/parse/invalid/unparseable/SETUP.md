# Scenario

**Feature**: `--idle-timeout=nope` is a parse error

```
ParseRunIdle(_, "nope") -> error (before Normalize)
stderr: Error: invalid value for --idle-timeout: nope
```

## Steps

1. Set `IdleTimeoutRaw=nope` (flag off — invalid string fails first).

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.ExitOnIdle = false
	req.IdleTimeoutRaw = "nope"
	req.Args = []string{"run", "--idle-timeout=nope"}
	return nil
}
```
