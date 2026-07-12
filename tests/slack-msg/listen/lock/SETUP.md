# Scenario

**Feature**: singleton lock prevents concurrent listeners

```
# explicit --lock-file PATH
slack-msg listen --lock-file PATH (holds) -> second -> another slack-msg is already running

# default path when flag omitted (~/.agent-pro/slack-msg.listen.lock under HOME)
slack-msg listen (no --lock-file) -> second -> same singleton error

# --no-lock disables singleton
slack-msg listen --no-lock -> second without "already running"
```

## Preconditions

- Leaves choose explicit `--lock-file`, product default (`UseDefaultLock` + `HomeDir`), or `--no-lock`.
- Daemon isolation: non-lock leaves outside this group auto-get a WorkDir lock file via harness.

## Steps

1. Isolate workdir and tokens.
2. Leaf configures lock mode and second-instance probe as needed.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.WorkDir = t.TempDir()
	prependListenTokens(req)
	req.ClearSlackEnv = true
	return nil
}
```
