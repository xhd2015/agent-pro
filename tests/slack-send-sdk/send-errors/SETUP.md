# Scenario

**Feature**: send path fails after config and resolution

```
slack-send -> PostMessage error -> stderr send failed: ... -> exit 1, no OK line
```

## Preconditions

- Uses slacktest with `SLACK_API_URL` for deterministic API errors.

## Steps

1. Grouping starts slacktest and sets `req.SlackAPIURL`.
2. Leaf supplies invalid channel or error-inducing args.

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
	req.ConfigInline = strings.Replace(req.ConfigInline, "xoxb-doctest-fake-token", "xoxb-slacktest-token", 1)
	return nil
}
```