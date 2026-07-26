# Scenario

**Feature**: `--topic add-missing-scope --help` same as help-then-topic

```
Caller -> slack-msg --topic add-missing-scope --help -> same topic body -> exit 0
```

## Steps

1. Args `["--topic", "add-missing-scope", "--help"]`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Args = []string{"--topic", "add-missing-scope", "--help"}
	return nil
}
```
