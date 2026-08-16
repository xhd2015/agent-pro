# Scenario

**Feature**: session history --help

```
slack-msg session history --help -> same as -h
```

## Steps

1. Args: session history --help.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Args = []string{"session", "history", "--help"}
	return nil
}
```
