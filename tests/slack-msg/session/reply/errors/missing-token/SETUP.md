# Scenario

**Feature**: session reply missing bot token/config

```
map entry without config_path; no --config / --token / SLACK_MSG_CONFIG
  -> bot token required or config error; exit 1
```

## Steps

1. Seed map entry with empty config_path.
2. --session-id + MESSAGE only.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	if err := seedSessionsJSON(t, req.HomeDir, []sessionMapEntry{{
		SessionID:  "slack-channel-C0ALE44K5J6",
		ChannelID:  slackTestChannelID,
		ThreadTS:   "1710000800.000100",
		ConfigPath: "",
		Kind:       "channel",
		ReplyMode:  "channel",
	}}); err != nil {
		return err
	}
	req.Args = []string{
		"session", "reply",
		"--session-id", "slack-channel-C0ALE44K5J6",
		"hello",
	}
	return nil
}
```
