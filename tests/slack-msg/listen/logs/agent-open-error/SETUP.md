# Scenario

**Feature**: agent open failure is logged (not silent)

```
inbound accepted -> mock agent exits 1 -> stderr error line
```

## Steps

1. Use failing mock agent-run.
2. Inject one app_mention; wait for launch log then assert error log.

```go
import (
	"path/filepath"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.WorkDir = t.TempDir()
	req.AgentLogPath = filepath.Join(req.WorkDir, "agent.log")
	req.MockAgentPath = writeMockAgentFail(t, req.WorkDir, req.AgentLogPath)
	req.WantAgentCalls = 1
	req.InjectEvents = []InjectedEvent{{
		Kind:    "app_mention",
		Channel: slackTestChannelID,
		User:    slackTestUserID,
		Text:    "<@" + slackTestBotUserID + "> fail open",
		TS:      "1710000610.000100",
	}}
	return nil
}
```
