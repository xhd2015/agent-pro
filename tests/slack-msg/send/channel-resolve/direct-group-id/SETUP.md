# Scenario

**Feature**: private group ID used as-is

```
slack-msg send --channel G024BE91L -> passthrough -> send OK
```

## Steps

1. Pass group ID directly.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Args = []string{
		"send",
		"--token", slackTestToken,
		"--channel", "G024BE91L",
		"direct G",
	}
	return nil
}
```
