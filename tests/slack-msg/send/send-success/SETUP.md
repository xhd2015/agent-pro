# Scenario

**Feature**: successful send via slacktest fake API

```
slack-msg send [options] MESSAGE -> SLACK_API_URL=slacktest -> PostMessage ok -> OK ts=... channel=... -> exit 0
```

## Preconditions

- `SLACK_API_URL` test hook in `cmd/slack-msg`.
- Session-scoped `slacktest` server with `conversations.list` + `chat.postMessage`.

## Steps

1. Start or reuse slacktest server; set `req.SlackAPIURL`.
2. Clear `SLACK_CONFIG` env; no `--config` unless leaf overrides.
3. Leaf narrows credential source and message text.
4. Assert three stdout lines; stderr empty; exit 0.

## Context

- Timestamp from slacktest is dynamic — use `__TS__: type=number`.

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
