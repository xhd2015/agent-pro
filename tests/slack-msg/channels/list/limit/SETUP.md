# Scenario

**Feature**: --limit caps printed channels after filter/sort

```
slack-msg channels list --limit 2 -> first two by name
```

## Preconditions

- Default slacktest conversations.list fixture.

## Steps

1. Attach default slacktest server.
2. Leaf sets `--limit`.

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
