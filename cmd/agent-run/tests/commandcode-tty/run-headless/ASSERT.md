## Expected

- Exit code 0.
- Stderr matches `commandcode-tty: session-N`.
- Stdout has 2+ JSON lines, first is `{"type":"message",...}` with non-empty text, last is `{"type":"done"}`.

## Exit Code

0

```go
import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/xhd2015/doctest/assert"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.ExitCode != 0 {
		t.Fatalf("expected exit 0, got %d\nstderr:\n%s\nstdout:\n%s", resp.ExitCode, resp.Stderr, resp.Stdout)
	}

	assert.Output(t, resp.Stderr, "" +
`<contains>
<regex>commandcode-tty:\s*session-\d+</regex>
</contains>`)

	lines := strings.Split(strings.TrimSpace(resp.Stdout), "\n")
	if len(lines) < 2 {
		t.Fatalf("expected >= 2 JSON lines, got %d\nstdout:\n%s", len(lines), resp.Stdout)
	}

	var msg map[string]any
	if err := json.Unmarshal([]byte(lines[0]), &msg); err != nil {
		t.Fatalf("first line not valid JSON: %v", err)
	}
	if msg["type"] != "message" {
		t.Fatalf("expected type=message, got %v", msg["type"])
	}
	text, _ := msg["text"].(string)
	if strings.TrimSpace(text) == "" {
		t.Fatalf("message text is empty")
	}

	var done map[string]any
	if err := json.Unmarshal([]byte(lines[len(lines)-1]), &done); err != nil {
		t.Fatalf("last line not valid JSON: %v", err)
	}
	if done["type"] != "done" {
		t.Fatalf("expected type=done, got %v", done["type"])
	}
}
```
