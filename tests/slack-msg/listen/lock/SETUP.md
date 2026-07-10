# Scenario

**Feature**: singleton lock prevents concurrent listeners

```
slack-msg listen (holds lock) -> second slack-msg listen -> another slack-msg is already running
```

## Preconditions

- Custom `--lock-file` in temp workdir for isolation.

## Steps

1. Start first instance as daemon with lock file.
2. Attempt second instance while first is running.

```go
import (
	"path/filepath"
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	req.WorkDir = t.TempDir()
	req.LockFile = filepath.Join(req.WorkDir, "slack-msg.lock")
	prependListenTokens(req)
	req.Daemon = true
	req.SecondInstance = true
	req.ClearSlackEnv = true
	return nil
}
```
