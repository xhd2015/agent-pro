# Scenario

**Feature**: session reply resolves session id + config from env

```
SLACK_MSG_SESSION_ID + SLACK_MSG_CONFIG + map entry + MESSAGE
  -> PostMessage channel-only -> OK
```

## Steps

1. Seed map entry.
2. Materialize config; set env vars (not flags).
3. Args: session reply MESSAGE only.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	if err := seedSessionsJSON(t, req.HomeDir, []sessionMapEntry{{
		SessionID: sessionReplyFixtureID,
		ChannelID: slackTestChannelID,
		ThreadTS:  "1710000800.000100",
		Kind:      "channel",
		ReplyMode: "channel",
	}}); err != nil {
		return err
	}
	req.ConfigFixture = "valid-config.json"
	if err := materializeConfig(t, d, req); err != nil {
		return err
	}
	req.Env = append(req.Env,
		"SLACK_MSG_SESSION_ID="+sessionReplyFixtureID,
		"SLACK_MSG_CONFIG="+req.ConfigPath,
	)
	req.Args = []string{"session", "reply", "env reply body"}
	return nil
}
```
