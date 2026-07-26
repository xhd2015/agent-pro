# Scenario

**Feature**: `send --help` shows help

```
Caller -> slack-msg send --help -> usage stdout -> exit 0
```

## Steps

1. Args `["send", "--help"]`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Args = []string{"send", "--help"}
	return nil
}
```
