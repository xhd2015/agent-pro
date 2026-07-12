# Scenario

**Feature**: second instance rejected while explicit lock held

```
first listen --lock-file PATH acquires lock -> second listen same PATH -> exit non-zero with singleton message
```

## Steps

1. Use isolated `--lock-file` under workdir.
2. Start first instance as daemon; attempt second while first is running.

```go
import (
	"path/filepath"
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	req.LockFile = filepath.Join(req.WorkDir, "slack-msg.lock")
	req.Daemon = true
	req.SecondInstance = true
	req.InjectEvents = nil
	req.WantAgentCalls = 0
	return nil
}
```
