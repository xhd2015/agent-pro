# Scenario

**Feature**: conversations.list API returns error

```
slack-msg channels list -> conversations.list ok=false -> channels failed:
```

## Steps

1. Use channels-fail slacktest server.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.ChannelsAPIFail = true
	req.Args = []string{
		"channels", "list",
		"--token", slackTestToken,
	}
	return nil
}
```
