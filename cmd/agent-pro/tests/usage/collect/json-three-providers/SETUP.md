# Scenario

**Feature**: `--json` prints one record per provider, and the same records land on disk

```
usage collect --json
stdout: {"schema":"agent-pro/usage-snapshot/v1","ts":"...","provider":"grok",...}
        ... codex ...
        ... commandcode ...
usages/grok/<today>/<time>-snapshot.jsonl   one line, the same record
```

## Preconditions

- Every provider answers: grok and codex from fixtures, Command Code from the fake
  API.
- One session per provider home, so the session block has something to report.

## Steps

1. Add `--json` so stdout is exactly the records, with no table or trailer.

```go
import (
"testing"

"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
req.Args = []string{"--json"}
return nil
}
```
