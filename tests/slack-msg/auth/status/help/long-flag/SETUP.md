# Scenario

**Feature**: `auth status --help` shows help

```
Caller -> slack-msg auth status --help -> usage stdout -> exit 0
```

## Steps

1. Args `["auth", "status", "--help"]`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Args = []string{"auth", "status", "--help"}
	return nil
}
```
