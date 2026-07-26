# Scenario

**Feature**: multi-message history chronological order

```
slack-msg history --token --channel C0... -> [oldest] then [newer] then [newest]
```

## Steps

1. Flags for token and channel ID.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Args = []string{
		"history",
		"--token", slackTestToken,
		"--channel", "C0ALE44K5J6",
	}
	return nil
}
```
