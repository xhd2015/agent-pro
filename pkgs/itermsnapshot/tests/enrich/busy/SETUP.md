# Scenario

**Feature**: busy pane agent attach depends on ResolveFromPID outcome

```
Idle=false -> ResolveFromPID(pid) -> Agent | soft miss
```

## Steps

1. Leaves inject busy Snapshot sessions.
2. Split on resolve hard hit kind, none, error, and PID source.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	// NoEnrich already false from enrich parent.
	return nil
}
```
