# Scenario

**Feature**: `channels search -h` shows help

```
Caller -> slack-msg channels search -h -> usage stdout -> exit 0
```

## Steps

1. Args `["channels", "search", "-h"]`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Args = []string{"channels", "search", "-h"}
	return nil
}
```
