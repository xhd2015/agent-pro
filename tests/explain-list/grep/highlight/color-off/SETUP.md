# Scenario

**Feature**: without --color, --grep filters but does not paint match spans

```
# same Docker session; explain list --grep docker (no --color, non-TTY)
-> session kept; stdout has no ANSI escapes
```

## Preconditions

- One matching session; harness stdout is a pipe (auto color off).

## Steps

1. Seed `highlightDockerSession`.
2. Args: `list --grep docker`.
3. Assert listed + no ANSI.

## Context

- Color-off path must not inject bold-red (or any) body SGR.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Args = []string{"list", "--grep", "docker"}
	req.EnvExtra = nil
	req.Sessions = []SessionSeed{highlightDockerSession()}
	return nil
}
```
