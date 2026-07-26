# Scenario

**Feature**: thread mode follow-up also uses RunInteractiveOpen (not send)

```
first message open run --session-id=…
  -> second message same thread -> again open run --session-id=… (not send)
```

## Steps

1. Inject two messages sharing thread_ts.
2. Expect two interactive-open launches (no legacy `send` argv).

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	threadTS := "1710000200.000100"
	req.WantAgentCalls = 2
	req.InjectEvents = []InjectedEvent{
		{
			Kind:    "app_mention",
			Channel: slackTestChannelID,
			Text:    "<@" + slackTestBotUserID + "> thread start",
			TS:      threadTS,
		},
		{
			Channel:  slackTestChannelID,
			Text:     "follow up in thread",
			TS:       "1710000200.000200",
			ThreadTS: threadTS,
		},
	}
	return nil
}
```
