# Scenario

**Feature**: default search is case-insensitive contains on channel name

```
slack-msg channels search QUERY -> name contains query (strip #) -> matching rows
```

## Preconditions

- Default slacktest conversations.list fixture.

## Steps

1. Attach default slacktest server.
2. Leaf supplies QUERY.

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
