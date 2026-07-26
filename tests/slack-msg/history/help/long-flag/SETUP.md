# Scenario

**Feature**: `history --help` shows help

```
Caller -> slack-msg history --help -> usage stdout -> exit 0
```

## Steps

1. Args `["history", "--help"]`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Args = []string{"history", "--help"}
	return nil
}
```
