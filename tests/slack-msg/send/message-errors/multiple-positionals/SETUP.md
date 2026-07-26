# Scenario

**Feature**: two positional messages rejected

```
Caller -> slack-msg send ... "one" "two" -> exactly one message required
```

## Steps

1. Append two message positionals after flags.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Args = []string{"send", "--token", slackTestToken, "--channel", "C0ALE44K5J6", "first", "second"}
	return nil
}
```
