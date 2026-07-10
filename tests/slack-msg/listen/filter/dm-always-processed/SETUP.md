# Scenario

**Feature**: DMs always processed regardless of requireMention

```
DM message (no @bot) with default requireMention -> agent invoked
```

## Steps

1. Inject DM `message` event without mention.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.WantAgentCalls = 1
	req.InjectEvents = []InjectedEvent{{
		Kind: "dm",
		Text: "private question",
		TS:   "1710000002.000100",
	}}
	return nil
}
```
