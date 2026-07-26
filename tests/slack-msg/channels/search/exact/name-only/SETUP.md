# Scenario

**Feature**: --exact general matches only general (not gen substring)

```
slack-msg channels search --exact general -> #general only
```

## Steps

1. Token, `--exact`, QUERY `general`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Args = []string{
		"channels", "search",
		"--token", slackTestToken,
		"--exact",
		"general",
	}
	return nil
}
```
