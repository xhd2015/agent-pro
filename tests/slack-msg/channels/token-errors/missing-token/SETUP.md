# Scenario

**Feature**: missing bot token for channels list

```
Caller -> slack-msg channels list -> bot token required
```

## Steps

1. No token flags or env.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Args = []string{"channels", "list"}
	return nil
}
```
