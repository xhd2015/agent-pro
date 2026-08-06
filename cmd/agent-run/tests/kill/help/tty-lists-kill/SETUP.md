# Scenario

**Feature**: `tty --help` lists the `kill` subcommand

```
agent-run tty --help -> contains kill (alongside status, attach, send, …)
```

## Steps

1. Set Args to `tty --help`.
2. Run Mode handle.
3. Assert exit 0 and stdout lists `kill`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Mode = "handle"
	req.Args = []string{"tty", "--help"}
	return nil
}
```
