# Scenario

**Feature**: `channels -h` shows help

```
Caller -> slack-msg channels -h -> usage stdout -> exit 0
```

## Steps

1. Args `["channels", "-h"]`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Args = []string{"channels", "-h"}
	return nil
}
```
