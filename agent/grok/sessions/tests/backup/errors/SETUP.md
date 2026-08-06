# Scenario

**Feature**: Backup fails closed — error and no usable payload

```
# invalid input / busy session / archive rules
seed (optional) -> sessions.Backup(...)
  -> error; no manifest.json / payload written to OutDir
```

## Preconditions

- Leaves under this branch expect a non-nil error from `Backup`.
- When `OutDir` is set, asserts verify no `manifest.json` / `payload/`.

## Steps

1. Leaf seeds the specific failure condition.
2. `Run` calls `Backup`.
3. Assert error message class and absence of payload.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	// Errors leaves configure themselves; keep empty live injectables by default.
	return nil
}
```
