# Scenario

**Feature**: successful session reply via map + config + slacktest

```
seed sessions.json + CapturePosts slacktest
  -> slack-msg session reply … MESSAGE
  -> PostMessage channel only (thread_ts empty) -> OK
```

## Preconditions

- Isolated `HomeDir` with map entry for fixed session id.
- `CapturePosts` captures postMessage params.
- Clear host Slack env.

## Steps

1. Isolate home; enable CapturePosts; clear Slack env.
2. Leaf seeds map entry and sets argv / env.
3. Assert OK line, channel post, empty thread_ts.

## Context

- Session id fixture: `slack-channel-C0ALE44K5J6` (stable channel key).
- Channel: `C0ALE44K5J6`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
const sessionReplyFixtureID = "slack-channel-C0ALE44K5J6"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.ClearSlackEnv = true
	req.CapturePosts = true
	if err := isolateHome(t, req); err != nil {
		return err
	}
	return nil
}
```
