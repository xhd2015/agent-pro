# Scenario

**Feature**: `#general` resolved via conversations.list

```
slack-msg send --channel "#general" -> API list -> C0ALE44K5J6 -> send OK
```

## Steps

1. Args with channel name including `#`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Args = []string{
		"send",
		"--token", slackTestToken,
		"--channel", "#general",
		"resolve hash",
	}
	return nil
}
```
