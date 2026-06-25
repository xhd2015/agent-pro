## Expected

- `subagent.Run` succeeds.
- `questions/` directory does **not** exist.
- Captured stdout does **not** contain `QUESTIONS` footer block.
- `events.jsonl` still created.

## Side Effects

- `progress/` may still exist when ProgressEnabled true.

## Errors

- None.

## Exit Code

N/A

```go
import (
	"os"
	"path/filepath"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	dir := req.SessionDir
	if dirExists(filepath.Join(dir, "questions")) {
		t.Fatalf("questions/ should not exist when QuestionsEnabled is false")
	}
	if containsQuestionsFooter(resp.Stdout) {
		t.Fatalf("stdout contains QUESTIONS footer when questions disabled:\n%s", resp.Stdout)
	}
	if _, err := os.Stat(filepath.Join(dir, "events.jsonl")); err != nil {
		t.Fatalf("events.jsonl missing: %v", err)
	}
}
```
