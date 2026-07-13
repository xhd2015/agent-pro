# Scenario

**Feature**: session info success (human / json / env session id)

```
seed map + messages.jsonl
  -> session info --session-id ID | SLACK_MSG_SESSION_ID
  -> human keys or --json with message_count + session_dir
```

## Preconditions

- Isolated HomeDir; clear Slack env.
- Fixture: 2 messages → message_count 2; empty dir.

## Steps

1. Isolate home; seed map + two log lines.
2. Leaf sets info flags or env.

```go
import "testing"

func sessionInfoFixtureEntry() sessionMapEntry {
	return sessionMapEntry{
		SessionID:          sessionInfoFixtureID,
		ChannelID:          slackTestChannelID,
		ThreadTS:           "1710000900.000100",
		ConfigPath:         "/tmp/slack-info-cfg.json",
		Dir:                "",
		Kind:               "channel",
		ReplyMode:          "channel",
		LastMessagePreview: "info preview",
		CreatedAt:          "2026-07-10T12:00:00Z",
		UpdatedAt:          "2026-07-13T08:00:00Z",
	}
}

func sessionInfoFixtureMessages() []sessionLogMessage {
	return []sessionLogMessage{
		{MessageID: "i1", TS: "1710000901.000100", User: "U1", Text: "alpha", Direction: "in"},
		{MessageID: "i2", TS: "1710000902.000200", User: "U2", Text: "beta", Direction: "out"},
	}
}

func Setup(t *testing.T, req *Request) error {
	req.ClearSlackEnv = true
	if err := isolateHome(t, req); err != nil {
		return err
	}
	if err := seedSessionsJSON(t, req.HomeDir, []sessionMapEntry{sessionInfoFixtureEntry()}); err != nil {
		return err
	}
	if err := seedMessagesJSONL(t, req.HomeDir, sessionInfoFixtureID, sessionInfoFixtureMessages()); err != nil {
		return err
	}
	return nil
}
```
