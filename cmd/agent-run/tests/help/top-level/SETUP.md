# Scenario

**Feature**: top-level `--help` lists subcommands and flags (including `kill`)

```
# L2 in-process (Mode handle): agentruncli.Handle(["--help"])
# same tokens as binary agent-run --help
```

## Preconditions

- No binary build: `req.Mode = "handle"` uses `pkgs/agentruncli.Handle`.

## Steps

1. Set Mode handle and args `--help`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Mode = "handle"
	req.Args = []string{"--help"}
	return nil
}
```