# Scenario

**Feature**: `-h` shows top-level help

```
Caller -> slack-msg -h -> usage stdout -> exit 0
```

## Steps

1. Args `["-h"]`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Args = []string{"-h"}
	return nil
}
```
