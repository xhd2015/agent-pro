# Scenario

**Feature**: session --help

```
slack-msg session --help -> same usage as -h
```

## Steps

1. Args: session --help.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Args = []string{"session", "--help"}
	return nil
}
```
