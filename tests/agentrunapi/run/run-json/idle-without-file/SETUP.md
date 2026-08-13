# Scenario

**Feature**: RunJSON does not treat TTY idle as done when the result file is empty

```
Launch no-op; production wait; no JSON written; short timeout
  -> error (timeout), not success
```

## Steps

1. Install Launch only (no Wait hook).
2. Do not write the result file.
3. Timeout 250ms so the leaf stays fast.

```go
import (
	"testing"
	"time"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.InstallLaunch = true
	req.PollInterval = 20 * time.Millisecond
	req.Timeout = 250 * time.Millisecond
	return nil
}
```
