# Scenario

**Feature**: `auth status -h` shows help

```
Caller -> slack-msg auth status -h -> usage stdout -> exit 0
```

## Steps

1. Args `["auth", "status", "-h"]`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Args = []string{"auth", "status", "-h"}
	return nil
}
```
