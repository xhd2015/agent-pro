# Scenario

**Feature**: send with channel ID flag

```
slack-msg send --token --channel C0ALE44K5J6 MESSAGE -> OK
```

## Steps

1. Channel ID instead of name.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Args = []string{
		"send",
		"--token", slackTestToken,
		"--channel", "C0ALE44K5J6",
		"custom text here",
	}
	return nil
}
```
