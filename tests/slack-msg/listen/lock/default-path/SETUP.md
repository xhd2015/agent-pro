# Scenario

**Feature**: default lock path when `--lock-file` omitted

```
# HOME isolated -> default ~/.agent-pro/slack-msg.listen.lock
first listen (no --lock-file / no --no-lock) holds default lock
  -> second same HOME -> another slack-msg is already running
  -> banner/startup shows default lock path
```

## Steps

1. Set `HomeDir` under workdir so product default path is isolated.
2. Leave lock flags unset (`UseDefaultLock`).
3. Start daemon + second instance.

```go
import (
	"path/filepath"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.HomeDir = filepath.Join(req.WorkDir, "home")
	req.UseDefaultLock = true
	req.Daemon = true
	req.SecondInstance = true
	req.InjectEvents = nil
	req.WantAgentCalls = 0
	return nil
}
```
