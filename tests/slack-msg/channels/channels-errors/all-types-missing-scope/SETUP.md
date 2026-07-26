# Scenario

**Feature**: all requested types missing_scope hard-fails with no successful type

```
slack-msg channels list --token (default public,private)
  -> public_channel missing_scope + private_channel missing_scope
  -> stderr channels failed: + missing_scope
  -> exit 1
```

## Steps

1. Use all-types-missing-scope mock.
2. Default multi-type list.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.ChannelsAllMissingScope = true
	req.Args = []string{
		"channels", "list",
		"--token", slackTestToken,
	}
	return nil
}
```
