# Scenario

**Feature**: 65-rune entry spills to `--env-file`

```go
import (
	"testing"
	"unicode/utf8"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.SessionID = "sess-env-over"
	entry := envEntryAtRunes(envFileSpillMinRunes + 1)
	if utf8.RuneCountInString(entry) != envFileSpillMinRunes+1 {
		t.Fatalf("fixture want %d runes; got %d", envFileSpillMinRunes+1, utf8.RuneCountInString(entry))
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
