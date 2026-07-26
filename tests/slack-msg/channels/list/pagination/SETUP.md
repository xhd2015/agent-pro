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
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.ClearSlackEnv = true
	req.ChannelsPaginated = true
	return nil
}
```
