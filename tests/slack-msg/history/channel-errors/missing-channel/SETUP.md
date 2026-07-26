# Scenario

**Feature**: missing channel for history

```
Caller -> slack-msg history --token TOK -> channel required
```

## Steps

1. Token only.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Args = []string{"history", "--token", slackTestToken}
	return nil
}
```
