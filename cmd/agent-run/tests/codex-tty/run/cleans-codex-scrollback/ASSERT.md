## Expected Output

```text
<contains>
ls output:
AGENTS.md
cmd
pkgs
</contains>
```

## Expected

- Exit code 0.
- Stdout contains useful final output lines from the Codex turn.
- Stdout and stored events do not contain raw terminal control fragments.
- Stdout and stored events do not contain Codex TUI chrome or startup/progress status.

## Exit Code

0

```go
import (
	"strings"
	"testing"

	"github.com/xhd2015/doctest/assert"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	assertSuccess(t, resp)

	assert.Output(t, resp.Stdout, `
<contains>
ls output:
AGENTS.md
cmd
pkgs
</contains>`)

	_, lines := findCodexTTYEventsJSONL(t, req.Home)
	events := strings.Join(lines, "\n")
	for _, useful := range []string{"ls output:", "AGENTS.md", "cmd", "pkgs"} {
		if !strings.Contains(resp.Stdout, useful) {
			t.Fatalf("stdout missing useful line %q:\n%s", useful, resp.Stdout)
		}
		if !strings.Contains(events, useful) {
			t.Fatalf("events missing useful line %q:\n%s", useful, events)
		}
	}

	assertNoRawCodexTranscript(t, "stdout", resp.Stdout)
	assertNoRawCodexTranscript(t, "events", events)
}

func assertNoRawCodexTranscript(t *testing.T, label, text string) {
	t.Helper()
	for _, forbidden := range []string{
		">4;0m",
		">7u",
		"╭",
		"╰",
		"│",
		"OpenAI Codex",
		"model: loading",
		"directory:",
		"Starting MCP servers",
		"Booting MCP",
		"Working",
		"› run ls",
	} {
		if strings.Contains(text, forbidden) {
			t.Fatalf("%s contains raw Codex TUI/control fragment %q:\n%s", label, forbidden, text)
		}
	}
}
```
