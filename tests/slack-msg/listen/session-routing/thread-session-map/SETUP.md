# Scenario

**Feature**: thread mode upserts durable sessions.json map entry

```
app_mention + --config PATH -> ~/.agent-pro/slack-local-bot/sessions.json
  entry: session_id, channel_id, thread_ts, config_path (abs), reply_mode=channel
```

## Steps

1. Isolate HOME; materialize config so listen has abs ConfigPath.
2. Inject thread message; wait for agent open.
3. Assert sessions.json entry fields.

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
	if err := withConfigArg(t, req, "valid-config.json", false); err != nil {
		return err
	}
	// Daemon defaultListenArgs appends --config from ConfigPath; avoid double via Args.
	// withConfigArg only inserts for non-ListenMode; ListenMode uses ConfigPath field.
	threadTS := "1710000720.000100"
	req.WantAgentCalls = 1
	req.InjectEvents = []InjectedEvent{{
		Kind:    "app_mention",
		Channel: slackTestChannelID,
		User:    slackTestUserID,
		Text:    "<@" + slackTestBotUserID + "> map me",
		TS:      threadTS,
	}}
	return nil
}
```
