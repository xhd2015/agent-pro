# Scenario

**Feature**: ignore bot-authored messages

```
bot user posts message -> filter drops -> no agent invocation
```

## Steps

1. Inject channel message with `user` = bot ID.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.WantAgentCalls = 0
	req.InjectEvents = []InjectedEvent{{
		Channel: slackTestChannelID,
		User:    slackTestBotUserID,
		Text:    "bot talking to itself",
		TS:      "1710000001.000100",
	}}
	return nil
}
```
