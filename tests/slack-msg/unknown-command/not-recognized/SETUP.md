# Scenario

**Feature**: unrecognized command name

```
Caller -> slack-msg not-a-command -> stderr + exit 1
```

## Steps

1. Args `["not-a-command"]`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Args = []string{"not-a-command"}
	return nil
}
```
