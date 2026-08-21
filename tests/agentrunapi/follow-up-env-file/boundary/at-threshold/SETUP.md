# Scenario

**Feature**: exactly 64-rune entry stays inline

```go
import (
	"testing"
	"unicode/utf8"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.SessionID = "sess-env-at"
	entry := envEntryAtRunes(envFileSpillMinRunes)
	if utf8.RuneCountInString(entry) != envFileSpillMinRunes {
		t.Fatalf("fixture want %d runes; got %d (%q)", envFileSpillMinRunes, utf8.RuneCountInString(entry), entry)
	}
	req.Env = []string{entry}
	dir, err := ensureSpillDir(t, d, "spill")
	if err != nil {
		return err
	}
	req.EnvSpillDir = dir
	return nil
}
```
