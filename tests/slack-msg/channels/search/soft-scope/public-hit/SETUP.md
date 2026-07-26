# Scenario

**Feature**: search finds public channel after soft-skipping private missing_scope

```
slack-msg channels search --token "#GENERAL"
  -> public hit #general; private skipped
  -> human line + stderr warning
  -> exit 0
```

## Steps

1. Token + QUERY `#GENERAL` (contains match; avoids false positive on `agent-pro-debug` which is private and soft-skipped anyway).

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Args = []string{
		"channels", "search",
		"--token", slackTestToken,
		"#GENERAL",
	}
	return nil
}
```
