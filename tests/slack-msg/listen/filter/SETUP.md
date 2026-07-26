# Scenario

**Feature**: inbound event filtering before agent dispatch

```
Socket Mode event -> slack-msg listen filter (mention/DM/allowFrom/bot-self/channel) -> agent-run or drop
```

## Preconditions

- Default `--require-mention` true unless leaf overrides.
- Daemon probe with slacktest + mock agent.

## Steps

1. Start daemon with filter-related flags.
2. Inject one Socket Mode `events_api` envelope.
3. Assert agent invoked or filtered per leaf.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.ClearSlackEnv = true
	prependListenTokens(req)
	req.Daemon = true
	return nil
}
```
