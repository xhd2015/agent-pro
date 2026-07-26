## Expected

- `subagent.Run` succeeds.
- `progress/` directory does **not** exist under session dir.
- `events.jsonl` and `messages.jsonl` exist.

## Side Effects

- `questions/` may still exist when QuestionsEnabled true.

## Errors

- None.

## Exit Code

N/A

```go
import (
	"os"
	"path/filepath"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	dir := req.SessionDir
	if dirExists(filepath.Join(dir, "progress")) {
		t.Fatalf("progress/ should not exist when ProgressEnabled is false")
	}
	for _, name := range []string{"events.jsonl", "messages.jsonl"} {
		if _, err := os.Stat(filepath.Join(dir, name)); err != nil {
			t.Fatalf("missing %s: %v", name, err)
		}
	}
}
```
