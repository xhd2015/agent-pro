## Preconditions
- Tests for `Config.SessionEnvVar`, `Config.SessionMetaField`, and `Config.DebugSessionEnv`.
- When set, these override the default env var / meta field / debug env names.
- When empty, defaults are used.

## Steps
1. Each sub-grouping node configures the relevant Config field(s).
2. Each leaf sets the custom or default value and verifies behavior via the public API.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
    _ = req.RoleName
    return nil
}
```
