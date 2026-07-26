# Scenario

**Feature**: allowFrom blocks unknown users

```
--allow-from specific user + message from other user -> filtered
```

## Steps

1. Restrict allowFrom to slackTestUserID.
2. Inject message from slackTestOtherUserID.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Args = []string{"--allow-from", slackTestUserID, "--no-require-mention"}
	req.WantAgentCalls = 0
	req.InjectEvents = []InjectedEvent{{
		Channel: slackTestChannelID,
		User:    slackTestOtherUserID,
		Text:    "blocked user says hi",
		TS:      "1710000005.000100",
	}}
	return nil
}
```
