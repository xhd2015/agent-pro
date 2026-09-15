# Scenario

**Feature**: codex usage survives the codex/usage endpoint being unavailable

```
delete fixture codex-usage.json
Collect -> codex backend-api/codex/usage fails -> backend-api/wham/usage -> 62% used
```

## Preconditions

- Inherits the root fixtures minus `codex-usage.json`, the file backing
  `backend-api/codex/usage`.
- `codex-wham.json` still answers `backend-api/wham/usage`, which reports usage in
  the same shape.

## Steps

1. Remove `codex-usage.json` so only the wham payload can answer.

```go
import (
"testing"

"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
delete(req.Fixtures, "codex-usage.json")
return nil
}
```
