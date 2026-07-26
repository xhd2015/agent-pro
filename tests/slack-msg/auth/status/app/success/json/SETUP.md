# Scenario

**Feature**: app status --json structured document

```
slack-msg auth status --app --app-token --json -> JSON app status
```

## Steps

1. Flags `--app`, `--app-token`, `--json`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Args = []string{
		"auth", "status",
		"--app",
		"--app-token", slackTestAppToken,
		"--json",
	}
	return nil
}
```
