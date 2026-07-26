# Scenario

**Feature**: missing bot token for history

```
Caller -> slack-msg history --channel C... -> bot token required
```

## Steps

1. Channel only.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Args = []string{"history", "--channel", "C0ALE44K5J6"}
	return nil
}
```
