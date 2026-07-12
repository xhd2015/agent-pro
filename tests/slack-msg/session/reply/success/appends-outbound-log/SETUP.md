# Scenario

**Feature**: successful session reply appends outbound line to messages.jsonl

```
reply success -> sessions/<id>/messages.jsonl gains direction=out + text
```

## Steps

1. Seed map + empty/inbound log optional.
2. Reply with --session-id + --token (skip config file path).
3. Assert outbound log line.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	if err := seedSessionsJSON(t, req.HomeDir, []sessionMapEntry{{
		SessionID: sessionReplyFixtureID,
		ChannelID: slackTestChannelID,
		ThreadTS:  "1710000800.000100",
		Kind:      "channel",
		ReplyMode: "channel",
	}}); err != nil {
		return err
	}
	// Seed one inbound line so append is ordered after it.
	if err := seedMessagesJSONL(t, req.HomeDir, sessionReplyFixtureID, []sessionLogMessage{{
		MessageID: "1710000800.000100",
		TS:        "1710000800.000100",
		User:      slackTestUserID,
		Text:      "prior inbound",
		Direction: "in",
	}}); err != nil {
		return err
	}
	req.Args = []string{
		"session", "reply",
		"--session-id", sessionReplyFixtureID,
		"--token", slackTestToken,
		"outbound reply text",
	}
	return nil
}
```
