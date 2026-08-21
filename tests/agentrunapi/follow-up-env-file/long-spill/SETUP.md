# Scenario

**Feature**: long env entry auto-spills all env to `--env-file`

```go
import (
	"testing"
	"unicode/utf8"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.SessionID = "sess-env-long"
	long := longPATHEntry()
	req.Env = []string{long}
	req.Open = true
	dir, err := ensureSpillDir(t, d, "spill")
	if err != nil {
		return err
	}
	req.EnvSpillDir = dir
	if utf8.RuneCountInString(long) <= envFileSpillMinRunes {
		t.Fatalf("long fixture must be >%d runes; got %d", envFileSpillMinRunes, utf8.RuneCountInString(long))
	}
	return nil
}
```
