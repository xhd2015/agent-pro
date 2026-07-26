# Scenario

**Feature**: listen startup banner after AuthTest

```
slack-msg listen -> AuthTest -> stderr/stdout banner
  (config path, team, bot identity, session-mode, require-mention, agent-runner, lock)
```

## Preconditions

- Daemon probe with slacktest auth.test / bots.info fixtures.
- Explicit `--lock-file` so banner lock line is deterministic.

## Steps

1. Start listen daemon with tokens; no inbound events required for identity banner.

```go
import (
	"path/filepath"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.ClearSlackEnv = true
	prependListenTokens(req)
	req.WorkDir = t.TempDir()
	req.LockFile = filepath.Join(req.WorkDir, "listen.lock")
	req.Daemon = true
	req.WantAgentCalls = 0
	return nil
}
```
