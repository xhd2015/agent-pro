# Scenario

**Feature**: PostMessage API returns error

```
slack-msg send -> chat.postMessage ok=false -> send failed:
```

## Steps

1. Use slacktest server with failing postMessage handler.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	apiURL, err := ensureSlackTestServerPostFail(t)
	if err != nil {
		return err
	}
	req.SlackAPIURL = apiURL
	req.Args = []string{
		"send",
		"--token", slackTestToken,
		"--channel", "C0ALE44K5J6",
		"post fail",
	}
	return nil
}
```
