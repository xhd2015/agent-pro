# Scenario

**Bug**: ask without dry-run on non-macOS must fail with a helpful error

```
non-darwin + no DEBUG_WITH_USER_DRY_RUN -> ask -> exit 2, stderr explains macOS requirement
```

## Preconditions

- v1 dialogs are macOS-only (`osascript`).
- Dry-run env vars are **not** set for this leaf.

## Steps

1. Run `ask` with default flags only (no `DEBUG_WITH_USER_DRY_RUN`).
2. On `darwin`, skip — this path is validated on Linux CI.

```go
import (
	"runtime"
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	if runtime.GOOS == "darwin" {
		t.Skip("non-mac error path is exercised on non-darwin CI; skip on macOS host")
	}
	// Ensure dry-run is off even if inherited from parent (non-mac is not under dry-run/).
	for i := 0; i < len(req.Env); i++ {
		if req.Env[i] == "DEBUG_WITH_USER_DRY_RUN=1" {
			req.Env = append(req.Env[:i], req.Env[i+1:]...)
			i--
		}
	}
	return nil
}
```
