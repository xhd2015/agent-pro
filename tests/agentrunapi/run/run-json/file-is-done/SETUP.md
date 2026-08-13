# Scenario

**Feature**: valid ResultFile ends production wait without a TTY

```
Launch writes JSON to ResultFile; Wait is production
  -> RunJSON returns; no timeout
```

## Steps

1. LaunchWritesResult; do not install Wait (production waitUntilDone).
2. Short poll interval.

```go
import (
	"testing"
	"time"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.InstallLaunch = true
	req.LaunchWritesResult = true
	req.WaitWriteJSON = `{"status":"ready"}`
	req.PollInterval = 10 * time.Millisecond
	req.Timeout = 2 * time.Second
	return nil
}
```
