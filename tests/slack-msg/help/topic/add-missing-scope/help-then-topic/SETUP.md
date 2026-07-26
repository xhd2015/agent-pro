# Scenario

**Feature**: `--help --topic add-missing-scope` prints grant guideline

```
Caller -> slack-msg --help --topic add-missing-scope -> topic body stdout -> exit 0
```

## Steps

1. Args `["--help", "--topic", "add-missing-scope"]`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Args = []string{"--help", "--topic", "add-missing-scope"}
	return nil
}
```
