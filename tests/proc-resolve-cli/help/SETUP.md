# Scenario

**Feature**: proc resolve help documents the resolve subcommand and flags

```
agent-pro proc resolve -h | agent-pro proc --help
  -> exit 0
  -> mentions resolve and --json
```

## Preconditions

- No test snapshot env required.
- Help is success (exit 0), not an error path.

## Steps

1. Leaves choose which help argv form to invoke.
2. Assert exit 0 and required tokens in stdout+stderr.

## Context

- Accept help on either `proc resolve -h` / `proc resolve --help` or
  `proc --help` as long as resolve and `--json` appear. Leaf locks one argv.

```go
import (
	"fmt"
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	// Help leaves only need the built binary; verify root Setup ran.
	if req.Bin == "" {
		return fmt.Errorf("help branch: req.Bin empty (root Setup must build agent-pro)")
	}
	// Leaves under help/ set req.Args to a help invocation.
	return nil
}
```

