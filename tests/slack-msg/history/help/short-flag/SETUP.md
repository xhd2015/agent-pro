# Scenario

**Feature**: `history -h` shows help

```
Caller -> slack-msg history -h -> usage stdout -> exit 0
```

## Steps

1. Args `["history", "-h"]`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Args = []string{"history", "-h"}
	return nil
}
```
