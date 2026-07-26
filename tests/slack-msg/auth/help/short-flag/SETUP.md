# Scenario

**Feature**: `auth -h` shows help

```
Caller -> slack-msg auth -h -> usage stdout -> exit 0
```

## Steps

1. Args `["auth", "-h"]`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Args = []string{"auth", "-h"}
	return nil
}
```
