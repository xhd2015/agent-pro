# Scenario

**Feature**: `--topic` without `--help` is an error

```
Caller -> slack-msg --topic add-missing-scope -> stderr requires --help -> exit 1
```

## Steps

1. Args `["--topic", "add-missing-scope"]` only (no `-h` / `--help`).

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Args = []string{"--topic", "add-missing-scope"}
	return nil
}
```
