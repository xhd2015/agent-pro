# Scenario

**Feature**: --prefix gen matches general

```
slack-msg channels search --prefix gen -> #general
```

## Steps

1. Token, `--prefix`, QUERY `gen`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Args = []string{
		"channels", "search",
		"--token", slackTestToken,
		"--prefix",
		"gen",
	}
	return nil
}
```
