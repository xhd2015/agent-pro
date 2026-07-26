# Scenario

**Feature**: missing channel

```
Caller -> slack-msg send --token TOK MESSAGE -> channel required
```

## Steps

1. Token + message only.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Args = []string{"send", "--token", slackTestToken, "Hello"}
	return nil
}
```
