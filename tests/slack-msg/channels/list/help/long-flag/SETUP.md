# Scenario

**Feature**: `channels list --help` shows help

```
Caller -> slack-msg channels list --help -> usage stdout -> exit 0
```

## Steps

1. Args `["channels", "list", "--help"]`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Args = []string{"channels", "list", "--help"}
	return nil
}
```
