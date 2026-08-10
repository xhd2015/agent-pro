# Scenario

**Feature**: top-level `--help` lists the `takeover` command

```
agent-run --help -> contains takeover as a command inventory line
```

## Steps

1. Set Args to `--help`.
2. Run Mode handle.
3. Assert exit 0 and stdout lists `takeover` as its own command line.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Mode = "handle"
	req.Args = []string{"--help"}
	return nil
}
```
