# Scenario

**Feature**: flags only without MESSAGE

```
Caller -> slack-msg send --token --channel (no message) -> message required
```

## Steps

1. Token + channel flags only.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Args = []string{"send", "--token", slackTestToken, "--channel", "C0ALE44K5J6"}
	return nil
}
```
