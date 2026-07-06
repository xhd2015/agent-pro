# Scenario

**Feature**: Detect reports persistent install state and live sudo probes

```
# drop-in + manifest check (no sudo -n for Installed)
Manager.Detect -> IsInstalled (FS) -> sudo -n true -> sudo -n <command> -> Status
```

## Steps

1. Set `Request.Operation = "detect"`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Operation = "detect"
	return nil
}
```