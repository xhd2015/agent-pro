# Scenario

**Feature**: --color + --grep wraps match spans in bold red, keeps original case

```
# Q "Using Docker Compose…"; explain list --grep docker --color
-> body contains \x1b[1;31mDocker\x1b[0m (not lowercased)
```

## Preconditions

- One session with `Docker` in the question body.

## Steps

1. Seed `highlightDockerSession`.
2. Args: `list --grep docker --color`.
3. Assert bold-red around `Docker`; filter kept the session.

## Context

- Label colors (Q cyan / A green) may also be present; match span is bold red.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Args = []string{"list", "--grep", "docker", "--color"}
	req.Sessions = []SessionSeed{highlightDockerSession()}
	return nil
}
```
