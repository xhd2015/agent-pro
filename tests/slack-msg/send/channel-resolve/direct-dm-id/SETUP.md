# Scenario

**Feature**: DM channel ID used as-is

```
slack-msg send --channel D024BE91L -> passthrough -> send OK
```

## Steps

1. Pass DM ID directly.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Args = []string{
		"send",
		"--token", slackTestToken,
		"--channel", "D024BE91L",
		"direct D",
	}
	return nil
}
```
