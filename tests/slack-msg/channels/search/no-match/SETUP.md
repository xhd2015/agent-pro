# Scenario

**Feature**: search with no matches exits 0 with empty results

```
slack-msg channels search no-such-channel -> empty / {"channels":[]} -> exit 0
```

## Preconditions

- Default slacktest conversations.list fixture.

## Steps

1. Attach default slacktest server.
2. Leaf supplies non-matching QUERY (and optional `--json`).

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
