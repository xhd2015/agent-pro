# Scenario

**Feature**: human channel lines sorted by name; archived excluded

```
API unsorted channels (incl. archived) -> slack-msg channels list -> name-sorted non-archived lines
```

## Preconditions

- Default slacktest conversations.list fixture (`slackTestChannels`).

## Steps

1. Attach default slacktest server.
2. Leaf runs `channels list` with token.

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
