# Scenario

**Feature**: DM inbound uses slack-dm-{userID} session key (not slack-channel-D…)

```
DM message from user W in channel D…
  -> agent-run --session-id=slack-dm-{userID}
  (not slack-channel-{DM channel id})
```

## Steps

1. Isolate HOME for optional sessions.json path asserts.
2. Inject one DM message (harness maps Kind `dm` → channel D… + message).
3. Assert launch session id is user-keyed.

```go
import (
	"path/filepath"
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	if req.WorkDir == "" {
		req.WorkDir = t.TempDir()
	}
	req.HomeDir = filepath.Join(req.WorkDir, "home")
	req.WantAgentCalls = 1
	req.InjectEvents = []InjectedEvent{{
		Kind: "dm",
		User: slackTestUserID,
		Text: "dm hello session key",
		TS:   "1710001100.000100",
	}}
	return nil
}
```
