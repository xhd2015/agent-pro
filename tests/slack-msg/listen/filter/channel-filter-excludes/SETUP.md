# Scenario

**Feature**: --channel filter excludes other channels

```
--channel C0ALE44K5J6 + event in C0OTHERCHAN -> filtered
```

## Steps

1. Pass `--channel` for general only.
2. Inject app_mention in other channel.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Args = []string{"--channel", slackTestChannelID}
	req.WantAgentCalls = 0
	req.InjectEvents = []InjectedEvent{{
		Kind:    "app_mention",
		Channel: "C0OTHERCHAN",
		Text:    "<@" + slackTestBotUserID + "> wrong channel",
		TS:      "1710000007.000100",
	}}
	return nil
}
```
