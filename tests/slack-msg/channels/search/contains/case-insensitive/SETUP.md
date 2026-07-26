# Scenario

**Feature**: contains match is case-insensitive; leading # stripped

```
slack-msg channels search --token "#GENERAL" -> matches #general
```

## Steps

1. Token + QUERY `#GENERAL` (hash + uppercase).
   - Note: QUERY is `#GENERAL` (not `#GEN`) so fixture name `agent-pro-debug`
     (which contains substring `gen`) does not false-positive under contains match.

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
