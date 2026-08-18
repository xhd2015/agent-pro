# Scenario

**Feature**: `--exit-on-idle --idle-timeout=-1s` is a parse error

```
ParseRunIdle(true, "-1s") -> error
stderr: Error: --idle-timeout must be a positive duration (got -1s)
```

## Steps

1. Set `ExitOnIdle=true`, `IdleTimeoutRaw=-1s`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.ExitOnIdle = true
	req.IdleTimeoutRaw = "-1s"
	req.Args = []string{"run", "--exit-on-idle", "--idle-timeout=-1s"}
	return nil
}
```
