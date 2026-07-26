# Scenario

**Feature**: multi-channel list sorted by name with member column

```
slack-msg channels list --token -> agent-pro-debug, general, random (excl. old-stuff)
```

## Steps

1. Flags for token only (default types; no --json).

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Args = []string{
		"channels", "list",
		"--token", slackTestToken,
	}
	return nil
}
```
