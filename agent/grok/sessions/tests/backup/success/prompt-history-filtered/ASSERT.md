## Expected

- Backup succeeds.
- `payload/sessions/<cwd_key>/prompt_history.session.jsonl` exists.
- Every JSON line has `session_id` equal to parent or child.
- Noise session id never appears.
- Exactly 3 lines (2 parent + 1 child from standard fixture).

## Errors

- None.

```go
import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertNoHarnessErr(t, err)
	dir := assertSuccessfulBackup(t, req, resp)

	ph := filepath.Join(dir, "payload", "sessions", req.CWDKey, "prompt_history.session.jsonl")
	assertFileExists(t, ph)

	f, err := os.Open(ph)
	if err != nil {
		t.Fatalf("open prompt extract: %v", err)
	}
	defer f.Close()

	sc := bufio.NewScanner(f)
	var n int
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" {
			continue
		}
		n++
		var obj map[string]any
		if err := json.Unmarshal([]byte(line), &obj); err != nil {
			t.Fatalf("prompt line %d not JSON: %v\n%s", n, err, line)
		}
		sid, _ := obj["session_id"].(string)
		if sid != req.SessionID && sid != req.ChildSessionID {
			t.Fatalf("prompt line %d session_id=%q not parent/child", n, sid)
		}
		if strings.Contains(line, req.PromptNoiseID) {
			t.Fatalf("noise id leaked into extract: %s", line)
		}
	}
	if err := sc.Err(); err != nil {
		t.Fatalf("scan: %v", err)
	}
	if n != 3 {
		t.Fatalf("prompt extract lines = %d, want 3", n)
	}
}
```
