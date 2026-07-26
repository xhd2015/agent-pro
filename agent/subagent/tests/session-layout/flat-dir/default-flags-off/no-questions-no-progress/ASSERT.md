## Expected

- `subagent.Run` succeeds.
- `events.jsonl` exists under `SessionLayout.Dir`.
- `questions/` directory does **not** exist.
- `progress/` directory does **not** exist.
- Stdout does **not** contain `QUESTIONS` footer.

## Side Effects

- `messages.jsonl` written at default path under `Dir` (MessagesPath unset uses default join).

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
	if dirExists(filepath.Join(dir, "questions")) {
		t.Fatalf("questions/ should not exist when QuestionsEnabled is zero")
	}
	if dirExists(filepath.Join(dir, "progress")) {
		t.Fatalf("progress/ should not exist when ProgressEnabled is zero")
	}
	if containsQuestionsFooter(resp.Stdout) {
		t.Fatalf("unexpected QUESTIONS footer:\n%s", resp.Stdout)
	}
	if _, err := os.Stat(filepath.Join(dir, "events.jsonl")); err != nil {
		t.Fatalf("events.jsonl missing: %v", err)
	}
}```
