## Expected

- `subagent.Run` succeeds.
- `messages.jsonl` does **not** exist under session dir.
- `events.jsonl` exists and is non-empty.

## Side Effects

- Other default artifacts may still be created (questions/, progress/).

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
	if fileExists(filepath.Join(dir, "messages.jsonl")) {
		t.Fatalf("messages.jsonl should not exist when MessagesPath is empty")
	}
	if data, err := os.ReadFile(filepath.Join(dir, "events.jsonl")); err != nil || len(data) == 0 {
		t.Fatalf("events.jsonl missing or empty: err=%v", err)
	}
}
```
