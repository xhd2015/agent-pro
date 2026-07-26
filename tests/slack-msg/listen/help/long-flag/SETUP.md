# Scenario

**Feature**: `listen --help` shows help

```
Caller -> slack-msg listen --help -> usage stdout -> exit 0
```

## Steps

1. Args `["--help"]`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Args = []string{"--help"}
	return nil
}
```
