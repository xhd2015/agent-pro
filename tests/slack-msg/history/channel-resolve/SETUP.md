# Scenario

**Feature**: history channel resolution (ID or API name)

```
slack-msg history --channel CH -> resolve -> conversations.history
```

## Preconditions

- slacktest conversations.list + history.

## Steps

1. Attach default slacktest.
2. Leaf varies channel input form.

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
