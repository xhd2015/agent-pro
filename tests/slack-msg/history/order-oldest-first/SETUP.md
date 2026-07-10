# Scenario

**Feature**: human history lines printed oldest→newest

```
API newest-first messages -> slack-msg history -> human lines oldest→newest
```

## Preconditions

- slacktest history handler returns `historyMessagesNewestFirst`.

## Steps

1. Attach default slacktest server.
2. Leaf runs history for known channel.

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
