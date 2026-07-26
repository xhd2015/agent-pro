# Scenario

**Feature**: multi-type list prints public channels when private returns missing_scope

```
slack-msg channels list --token
  -> public ok + private missing_scope
  -> human public lines + stderr soft warning
  -> exit 0
```

## Steps

1. Default types (`public,private`); token only; no `--json`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Args = []string{
		"channels", "list",
		"--token", slackTestToken,
	}
	return nil
}
```
