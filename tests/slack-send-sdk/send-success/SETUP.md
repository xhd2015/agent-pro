# Scenario

**Feature**: successful send via slacktest fake API

```
slack-send -> SLACK_API_URL=slacktest -> PostMessage ok -> OK ts=... channel=... -> exit 0
```

## Preconditions

- `github.com/slack-go/slack` with `SLACK_API_URL` test hook in `main.go` (implementer).
- Session-scoped `slacktest` server URL from `ensureSlackTestServer`.

## Steps

1. Start or reuse slacktest server; set `req.SlackAPIURL`.
2. Use inline config with `botToken: "xoxb-slacktest-token"`.
3. Leaf narrows args (default, channel name, custom text).
4. Assert three stdout lines; stderr empty; exit 0.

## Context

- Expected RED until SDK refactor and `SLACK_API_URL` hook land.
- Timestamp from slacktest is dynamic — use `__TS__` placeholder.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.WriteGoMod = true
	req.UseRepoConfig = false
	apiURL, err := ensureSlackTestServer(t)
	if err != nil {
		return err
	}
	req.SlackAPIURL = apiURL
	req.ConfigInline = readFixture(t, "valid-config.json")
	// slacktest accepts any non-empty token; override fixture token for clarity.
	req.ConfigInline = strings.Replace(req.ConfigInline, "xoxb-doctest-fake-token", "xoxb-slacktest-token", 1)
	return nil
}
```