# Scenario

**Feature**: `send -h` shows help

```
Caller -> slack-msg send -h -> usage stdout -> exit 0
```

## Steps

1. Args `["send", "-h"]`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Args = []string{"send", "-h"}
	return nil
}
```
