## Preconditions
- Input is `"  5m  "` — whitespace should be trimmed before parsing.
- 5m is ≥ 1m but < 10m, so a warning on stderr is expected.

## Steps
1. Set `req.Input` to `"  5m  "`.
2. `ParseTimeoutDuration` trims whitespace, parses as 5m, prints warning.

```go
import (
    "testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
    req.Input = "  5m  "
    return nil
}
```
