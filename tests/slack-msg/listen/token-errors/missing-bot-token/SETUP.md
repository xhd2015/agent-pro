# Scenario

**Feature**: missing bot token for listen

```
Caller -> slack-msg listen --app-token ... -> bot token required
```

## Steps

1. Provide app token only.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.AppToken = slackTestAppToken
	return nil
}
```
