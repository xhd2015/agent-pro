# Scenario

**Feature**: two separate channel messages share one stable session id

```
app_mention ts=T1 in channel C + app_mention ts=T2 in same C (different roots)
  -> both agent-run launches use --session-id=slack-channel-{C}
```

## Steps

1. Inject two channel app_mentions with different timestamps (no shared thread_ts).
2. Expect two interactive-open launches with the same stable channel session id.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.WantAgentCalls = 2
	req.InjectEvents = []InjectedEvent{
		{
			Kind:    "app_mention",
			Channel: slackTestChannelID,
			User:    slackTestUserID,
			Text:    "<@" + slackTestBotUserID + "> first channel msg",
			TS:      "1710001000.000100",
		},
		{
			Kind:    "app_mention",
			Channel: slackTestChannelID,
			User:    slackTestUserID,
			Text:    "<@" + slackTestBotUserID + "> second channel msg",
			TS:      "1710001000.000200",
		},
	}
	return nil
}
```
