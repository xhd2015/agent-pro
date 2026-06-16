## Preconditions

- The `agent/event/print` package exposes `FormatTraceLine(line string) string`.
- The `agent/event/print` package provides a `Coalescer` struct with `ShouldSkip`.
- The integration feeds raw JSON lines through: parse JSON → check coalescer → format → collect output.

## Steps

1. For each raw JSON line in `req.Lines`:
   a. Trim whitespace; skip empty/non-JSON lines.
   b. Unmarshal into `types.AgentEvent`.
   c. If the event is an `ActionMessage` and `Coalescer.ShouldSkip` returns true, append empty string (suppressed).
   d. Otherwise, call `print.FormatTraceLine(line)` and append the result.
2. Return `Response.Output` as the collected formatted lines (empty string = suppressed).

## Context

- `Request.Lines` is the sequence of raw JSON strings to process.
- `Response.Output[i]` is the formatted output for `Lines[i]`, or `""` if suppressed.

```go
import (
	"encoding/json"
	"strings"
	"testing"

	print "github.com/xhd2015/agent-pro/agent/event/print"
	types "github.com/xhd2015/agent-pro/agent/event/types"
)

type Request struct {
	Lines []string
}

type Response struct {
	Output []string
}

func Setup(t *testing.T, req *Request) error {
	_ = assertOutputNotEmpty
	_ = assertOutputEmpty
	return nil
}

func assertOutputNotEmpty(t *testing.T, idx int, output []string, msg string) {
	t.Helper()
	if output[idx] == "" {
		t.Fatalf("output[%d] must be non-empty: %s", idx, msg)
	}
}

func assertOutputEmpty(t *testing.T, idx int, output []string, msg string) {
	t.Helper()
	if output[idx] != "" {
		t.Fatalf("output[%d] must be empty: %s", idx, msg)
	}
}

func assertOutputContains(t *testing.T, idx int, output []string, substr string) {
	t.Helper()
	if !strings.Contains(output[idx], substr) {
		t.Fatalf("output[%d] must contain %q, got: %q", idx, substr, output[idx])
	}
}

func Run(t *testing.T, req *Request) (*Response, error) {
	var c print.Coalescer
	output := make([]string, len(req.Lines))
	for i, line := range req.Lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || !strings.HasPrefix(trimmed, "{") {
			output[i] = ""
			continue
		}
		var ev types.AgentEvent
		if err := json.Unmarshal([]byte(trimmed), &ev); err == nil && ev.Type == types.ActionMessage && c.ShouldSkip(ev) {
			output[i] = "" // suppressed by coalescer
			continue
		}
		output[i] = print.FormatTraceLine(line)
	}
	return &Response{Output: output}, nil
}
```
