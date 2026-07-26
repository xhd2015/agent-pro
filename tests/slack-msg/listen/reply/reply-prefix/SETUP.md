# Scenario

**Feature**: --reply-prefix prepended to Slack reply (stateless capture)

```
--session-mode stateless --reply-prefix [bot] + agent stdout
  -> PostMessage text starts with prefix; thread_ts = message ts
```

## Steps

1. Pass `--session-mode stateless` and `--reply-prefix [bot]`.
2. Inject processable app_mention.
3. Wait for PostMessage capture (`WantPosts = 1`).

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Args = append(req.Args, "--session-mode", "stateless", "--reply-prefix", "[bot]")
	req.WantPosts = 1
	req.InjectEvents = []InjectedEvent{{
		Kind:    "app_mention",
		Channel: slackTestChannelID,
		Text:    "<@" + slackTestBotUserID + "> prefix please",
		TS:      "1710000500.000100",
	}}
	return nil
}
```
