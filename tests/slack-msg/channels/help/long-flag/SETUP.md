# Scenario

**Feature**: `channels --help` shows help

```
Caller -> slack-msg channels --help -> usage stdout -> exit 0
```

## Steps

1. Args `["channels", "--help"]`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Args = []string{"channels", "--help"}
	return nil
}
```
