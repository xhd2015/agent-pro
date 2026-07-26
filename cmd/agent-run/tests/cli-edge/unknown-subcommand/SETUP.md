# Scenario

**Feature**: unknown subcommand exits with code 1

```
# L2 in-process: agentruncli.Handle(["not-a-real-command"]) → error → exit 1
```

## Preconditions

- No binary build: `req.Mode = "handle"` uses `pkgs/agentruncli.Handle`.

## Steps

1. Set Mode handle and args `not-a-real-command`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Mode = "handle"
	req.Args = []string{"not-a-real-command"}
	return nil
}
```