# Scenario

**Feature**: successful app auth status via apps.connections.open

```
slack-msg auth status --app --app-token -> SLACK_API_URL=slacktest -> connections.open ok
```

## Preconditions

- Default slacktest includes `/apps.connections.open`.

## Steps

1. Start or reuse default slacktest; set `req.SlackAPIURL`.
2. Clear Slack env; leaf supplies `--app` and app credentials.

## Context

- Fixed note: `app-level token (Socket Mode / connections); not used for channels/send/history`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.ClearSlackEnv = true
	apiURL, err := ensureSlackTestServer(t)
	if err != nil {
		return err
	}
	req.SlackAPIURL = apiURL
	return nil
}
```
