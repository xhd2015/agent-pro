# Scenario

**Feature**: `--idle-timeout=2s` without `--exit-on-idle` is parse-OK

```
ParseRunIdle(false, "2s") -> enabled=false, no error (silent no-op; no TTY)
```

## Steps

1. Set `ExitOnIdle=false`, `IdleTimeoutRaw=2s`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.ExitOnIdle = false
	req.IdleTimeoutRaw = "2s"
	req.Args = []string{"run", "--idle-timeout=2s"}
	return nil
}
```
