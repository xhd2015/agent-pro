# Scenario

**Feature**: --exact matches channel name exactly (after # strip)

```
slack-msg channels search --exact QUERY -> exact name only
```

## Preconditions

- Default slacktest conversations.list fixture.

## Steps

1. Attach default slacktest server.
2. Leaf sets `--exact` and QUERY.

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
