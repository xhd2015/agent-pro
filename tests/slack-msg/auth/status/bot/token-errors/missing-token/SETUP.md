# Scenario

**Feature**: missing bot token for auth status

```
Caller -> slack-msg auth status -> bot token required
```

## Steps

1. Args `["auth", "status"]` only.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Args = []string{"auth", "status"}
	return nil
}
```
