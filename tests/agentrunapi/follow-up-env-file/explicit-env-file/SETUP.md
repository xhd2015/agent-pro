# Scenario

**Feature**: explicit `EnvFile` is emitted; no auto-spill rewrite

```go
import (
	"strings"
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.SessionID = "sess-env-explicit"
	req.Open = true
	// Even with a long Env list, explicit EnvFile wins and must not auto-spill.
	req.Env = []string{longPATHEntry()}
	given, err := writeCaseFile(t, d, "given/path.env", "FROM_FILE=1\n")
	if err != nil {
		return err
	}
	req.EnvFile = given
	dir, err := ensureSpillDir(t, d, "spill")
	if err != nil {
		return err
	}
	req.EnvSpillDir = dir
	_ = strings.TrimSpace
	return nil
}
```
