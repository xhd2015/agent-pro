# Scenario

**Feature**: paginated conversations.list pages are merged before sort/filter

```
page1 + page2 cursor -> merge -> exclude archived -> sort by name
```

## Preconditions

- Paginated slacktest server (`ChannelsPaginated`).

## Steps

1. Enable paginated server mode.
2. Leaf runs `channels list`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.ClearSlackEnv = true
	req.ChannelsPaginated = true
	return nil
}
```
