# Scenario

**Feature**: --limit 2 yields two name-sorted lines

```
slack-msg channels list --limit 2 -> agent-pro-debug, general
```

## Steps

1. Flags for token and `--limit 2`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Args = []string{
		"channels", "list",
		"--token", slackTestToken,
		"--limit", "2",
	}
	return nil
}
```
