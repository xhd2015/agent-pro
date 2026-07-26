# Scenario

**Feature**: unknown channel name not in conversations.list

```
slack-msg send --channel #missing -> channel not found -> send failed
```

## Steps

1. Use slacktest default server (no `missing` channel in list).

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	apiURL, err := ensureSlackTestServer(t)
	if err != nil {
		return err
	}
	req.SlackAPIURL = apiURL
	req.Args = []string{
		"send",
		"--token", slackTestToken,
		"--channel", "#missing-channel",
		"nowhere",
	}
	return nil
}
```
