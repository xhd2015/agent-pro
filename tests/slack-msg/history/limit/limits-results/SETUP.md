# Scenario

**Feature**: --limit 2 yields two chronological lines

```
slack-msg history --limit 2 -> newest two from API -> print oldest of those first
```

## Steps

1. Flags for token, channel, `--limit 2`.

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
		"--limit", "2",
	}
	return nil
}
```
