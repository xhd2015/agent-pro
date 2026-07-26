# Scenario

**Feature**: bot status --json structured document

```
slack-msg auth status --token --json -> JSON status (no raw token)
```

## Steps

1. Flags `--token` and `--json`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Args = []string{
		"auth", "status",
		"--token", slackTestToken,
		"--json",
	}
	return nil
}
```
