# Scenario

**Feature**: successful bot auth status via slacktest auth.test

```
slack-msg auth status -> SLACK_API_URL=slacktest -> auth.test ok -> kind bot status
```

## Preconditions

- Session-scoped slacktest with overridden `/auth.test` (includes `bot_id`).

## Steps

1. Start or reuse default slacktest server; set `req.SlackAPIURL`.
2. Clear Slack env; leaf supplies `--token` and/or `--config`.
3. Assert human or JSON stdout; no raw full token; exit 0.

## Context

- auth.test fixture fields: team SlackTest Team / T024BE7LD, user Egon Spengler / W012A3CDE,
  bot_id B0TESTBOTID, url https://localhost.localdomain/.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.ClearSlackEnv = true
	apiURL, err := ensureSlackTestServer(t)
	if err != nil {
		return err
	}
	req.SlackAPIURL = apiURL
	return nil
}
```
