# Scenario

**Feature**: empty session id is rejected

```
RelocateCWD("", existing-target, {GrokHome}) -> error
```

## Preconditions

- Target directory may exist; validation of sessionID happens first.
- `req.SessionID` is the empty string.

## Steps

1. Create a target directory so target validation is not the only failure mode.
2. Set `req.SessionID = ""`.

```go
import (
	"path/filepath"
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	target := filepath.Join(req.TempDir, "ws-target")
	mustMkdir(t, target)
	req.TargetDir = absPath(t, target)
	req.SessionID = ""
	return nil
}
```
