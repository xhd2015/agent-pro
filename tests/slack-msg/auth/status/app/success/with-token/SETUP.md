# Scenario

**Feature**: app status with --app-token and no config

```
slack-msg auth status --app --app-token -> Using config from: (none) -> kind app
```

## Steps

1. Flags `--app` and `--app-token`.

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
	}
	return nil
}
```
