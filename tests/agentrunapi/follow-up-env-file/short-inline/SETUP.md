# Scenario

**Feature**: short env stays inline as `-e` (no `--env-file`)

```go
import (
	"testing"
	"unicode/utf8"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.SessionID = "sess-env-short"
	req.Env = []string{"NO_COLOR=1"}
	req.Open = true
	dir, err := ensureSpillDir(t, d, "spill")
	if err != nil {
		return err
	}
	req.EnvSpillDir = dir
	for _, e := range req.Env {
		if utf8.RuneCountInString(e) > envFileSpillMinRunes {
			t.Fatalf("short fixture entry %q must be ≤%d runes", e, envFileSpillMinRunes)
		}
	}
	return nil
}
```
