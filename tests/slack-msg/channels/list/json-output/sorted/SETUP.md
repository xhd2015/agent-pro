# Scenario

**Feature**: JSON document with channels sorted by name

```
slack-msg channels list --json --token -> sorted channels array
```

## Steps

1. Flags for token and `--json`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Args = []string{
		"channels", "list",
		"--token", slackTestToken,
		"--json",
	}
	return nil
}
```
