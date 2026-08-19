---
label: e2e
---

## Expected

- Exit code 0.
- Stderr contains `grok-tty: grok session a1111111-1111-4111-8111-111111111111`.
- Stdout or `events.jsonl` contains streamed assistant marker `AUTO_HOME_STREAM_MARKER`.
- Assistant output does **not** treat orchestrator stderr `GROK_HOME=` as assistant text
  (no `GROK_HOME=` in stdout; events lack scrollback-only pollution).

## Exit Code

0

```go
import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	assertSuccess(t, resp)
	assertStderrGrokSession(t, resp.Stderr, autoHomeUUID)

	combined := resp.Stdout + resp.Stderr
	if !strings.Contains(combined, "AUTO_HOME_STREAM_MARKER") {
		t.Fatalf("missing streamed marker in output; stdout:\n%s\nstderr:\n%s", resp.Stdout, resp.Stderr)
	}
	if strings.Contains(resp.Stdout, "GROK_HOME=") {
		t.Fatalf("stdout polluted with scrollback GROK_HOME= line:\n%s", resp.Stdout)
	}

	_, lines := findGrokTTYEventsJSONL(t, req.Home)
	for _, line := range lines {
		if strings.Contains(line, "GROK_HOME=") {
			t.Fatalf("events.jsonl polluted with GROK_HOME= scrollback: %s", line)
		}
	}
}

func findGrokTTYEventsJSONL(t *testing.T, home string) (string, []string) {
	t.Helper()
	root := filepath.Join(home, "sessions")
	var found string
	var lines []string
	_ = filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		if info.Name() == "events.jsonl" {
			found = path
			data, readErr := os.ReadFile(path)
			if readErr != nil {
				t.Fatalf("read %s: %v", path, readErr)
			}
			for _, line := range strings.Split(string(data), "\n") {
				if strings.TrimSpace(line) != "" {
					lines = append(lines, line)
				}
			}
		}
		return nil
	})
	return found, lines
}
```