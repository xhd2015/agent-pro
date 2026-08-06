## Expected

- No error.
- `Dir` is non-empty and ends with the session id path segment.
- `Files` includes basenames `summary.json`, `updates.jsonl`, `signals.json`.
- Each listed file has `Size` > 0, non-zero/non-empty `Path` containing its
  `Name`, and `Mtime` not zero.

## Errors

- None.

```go
import (
	"path/filepath"
	"strings"
	"testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertNoHarnessErr(t, err)
	assertNoError(t, resp)

	if resp.Dir == "" {
		t.Fatal("Dir empty")
	}
	if filepath.Base(resp.Dir) != req.SessionID {
		t.Fatalf("Dir base = %q, want session id %q (dir=%q)", filepath.Base(resp.Dir), req.SessionID, resp.Dir)
	}

	byName := fileNames(resp.Files)
	for _, want := range []string{"summary.json", "updates.jsonl", "signals.json"} {
		f, ok := byName[want]
		if !ok {
			t.Fatalf("missing file %q in %+v", want, resp.Files)
		}
		if f.Size <= 0 {
			t.Fatalf("%s Size = %d, want > 0", want, f.Size)
		}
		if f.Path == "" || !strings.Contains(f.Path, want) {
			t.Fatalf("%s Path = %q", want, f.Path)
		}
		if f.Mtime.IsZero() {
			t.Fatalf("%s Mtime is zero", want)
		}
	}
}
```
