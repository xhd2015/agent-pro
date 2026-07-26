# Scenario

**Feature**: missing bot token

```
Caller -> slack-msg send --channel C... MESSAGE -> bot token required
```

## Steps

1. Channel + message only.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Args = []string{"send", "--channel", "C0ALE44K5J6", "Hello"}
	return nil
}
```
