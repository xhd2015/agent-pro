# Scenario

**Feature**: session reply with --session-id and --config (map channel binding)

```
sessions.json entry + --session-id + --config + MESSAGE
  -> PostMessage(channel_id) without thread_ts -> OK
```

## Steps

1. Seed map entry with channel_id (config_path may be empty; use --config).
2. Args: session reply --session-id ID --config PATH MESSAGE.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	if err := seedSessionsJSON(t, req.HomeDir, []sessionMapEntry{{
		SessionID:  sessionReplyFixtureID,
		ChannelID:  slackTestChannelID,
		ThreadTS:   "1710000800.000100",
		ConfigPath: "",
		Kind:       "channel",
		ReplyMode:  "channel",
	}}); err != nil {
		return err
	}
	req.ConfigFixture = "valid-config.json"
	if err := materializeConfig(t, d, req); err != nil {
		return err
	}
	req.Args = insertConfigAfterSubcommand([]string{
		"session", "reply",
		"--session-id", sessionReplyFixtureID,
		"session reply body",
	}, req.ConfigPath)
	return nil
}
```
