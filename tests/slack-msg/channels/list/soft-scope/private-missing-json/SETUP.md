# Scenario

**Feature**: multi-type list --json soft-skips private missing_scope

```
slack-msg channels list --json --token
  -> public ok + private missing_scope
  -> JSON public only + stderr soft warning
  -> exit 0
```

## Steps

1. Default types; `--json`; token.

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
