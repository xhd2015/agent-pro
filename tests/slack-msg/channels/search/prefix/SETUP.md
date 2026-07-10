# Scenario

**Feature**: --prefix matches channel name by prefix (after # strip)

```
slack-msg channels search --prefix QUERY -> prefix match only
```

## Preconditions

- Default slacktest conversations.list fixture.

## Steps

1. Attach default slacktest server.
2. Leaf sets `--prefix` and QUERY.

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
