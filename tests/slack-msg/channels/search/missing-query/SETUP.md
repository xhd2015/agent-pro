# Scenario

**Feature**: search without QUERY is rejected

```
Caller -> slack-msg channels search --token TOK -> stderr query required -> exit 1
```

## Preconditions

- No positional QUERY.

## Steps

1. Clear Slack env vars.
2. Leaf provides token only (no QUERY).

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.ClearSlackEnv = true
	req.SlackAPIURL = ""
	return nil
}
```
