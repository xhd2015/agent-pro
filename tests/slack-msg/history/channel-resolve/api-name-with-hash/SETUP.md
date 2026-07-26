# Scenario

**Feature**: history channel name with hash resolved via API

```
slack-msg history --channel #general -> resolve C0ALE44K5J6 -> history
```

## Steps

1. Channel name with `#`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Args = []string{
		"history",
		"--token", slackTestToken,
		"--channel", "#general",
	}
	return nil
}
```
