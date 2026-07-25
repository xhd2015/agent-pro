# Scenario

**Feature**: agent-pro proc resolve runs ResolveFromPID and prints result

```
agent-pro proc resolve <pid> [--json] …
  -> library resolve (+ optional FormatTree / enrich)
doctest <- exit + stdout/stderr
```

## Preconditions

- json-hit installs `AGENT_PRO_PROCRESOLVE_TEST_SNAPSHOT` via `req.Snapshot`.
- unknown-pid uses a pid that does not exist on the machine (and no snapshot).

## Steps

1. Leaves set Args and optional Snapshot.
2. Assert exit code and output shape.

## Context

- Human mode (not covered as a required leaf) may print FormatTree; JSON mode
  must not depend on tree glyphs.

```go
import (
	"fmt"
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	if req.Bin == "" {
		return fmt.Errorf("resolve branch: req.Bin empty (root Setup must build agent-pro)")
	}
	// Leaves set Args and optional Snapshot (json-hit).
	return nil
}
```

