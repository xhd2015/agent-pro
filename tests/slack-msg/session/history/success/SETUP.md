# Scenario

**Feature**: session history success from local messages.jsonl

```
seed map + messages.jsonl under HOME
  -> slack-msg session history --session-id ID
  -> chronological human or --json
```

## Preconditions

- Isolated HomeDir; clear Slack env.
- Session id fixture shared with reply leaves.

## Steps

1. Isolate home; seed map + three log lines.
2. Leaf sets history flags.

## Context

- Log fixture ids: `m1`, `m2`, `m3` (ts order oldest→newest already).

```go
import "testing"

const sessionHistoryFixtureID = "slack-channel-C0ALE44K5J6"

func sessionHistoryFixtureMessages() []sessionLogMessage {
	return []sessionLogMessage{
		{MessageID: "m1", TS: "1710000901.000100", User: "U1", Text: "first", Direction: "in"},
		{MessageID: "m2", TS: "1710000902.000200", User: "U2", Text: "second", Direction: "out"},
		{MessageID: "m3", TS: "1710000903.000300", User: "U1", Text: "third", Direction: "in"},
	}
}

func Setup(t *testing.T, req *Request) error {
	req.ClearSlackEnv = true
	if err := isolateHome(t, req); err != nil {
		return err
	}
	if err := seedSessionsJSON(t, req.HomeDir, []sessionMapEntry{{
		SessionID: sessionHistoryFixtureID,
		ChannelID: slackTestChannelID,
		ThreadTS:  "1710000900.000100",
		Kind:      "channel",
		ReplyMode: "channel",
	}}); err != nil {
		return err
	}
	if err := seedMessagesJSONL(t, req.HomeDir, sessionHistoryFixtureID, sessionHistoryFixtureMessages()); err != nil {
		return err
	}
	return nil
}
```
