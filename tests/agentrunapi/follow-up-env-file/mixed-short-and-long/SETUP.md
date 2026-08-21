# Scenario

**Feature**: short + long entries spill **all** into one `--env-file`

```go
import (
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.SessionID = "sess-env-mixed"
	req.Env = []string{"NO_COLOR=1", longPATHEntry(), "FOO=bar"}
	req.Open = true
	dir, err := ensureSpillDir(t, d, "spill")
	if err != nil {
		return err
	}
	req.EnvSpillDir = dir
	return nil
}
```
