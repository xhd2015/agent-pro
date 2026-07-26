# Scenario

**Feature**: missing QUERY for channels search

```
Caller -> slack-msg channels search --token TOK -> query required
```

## Steps

1. Token only; no positional QUERY.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Args = []string{
		"channels", "search",
		"--token", slackTestToken,
	}
	return nil
}
```
