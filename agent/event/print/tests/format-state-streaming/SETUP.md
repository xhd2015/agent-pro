## Preconditions
- `print.FormatState.FormatLine` coalesces consecutive streaming events for trace display.
- Grok `thought` streaming produces multiple `ActionThink` AgentEvent lines (one per token).
- `traceSession` loops events through `FormatState`; this harness mirrors that loop.

## Steps
1. Set `req.Lines` to a slice of AgentEvent JSONL strings (simulating events.jsonl).
2. Feed each line through `FormatState.FormatLine`, collecting headers and bodies.
3. Call `Flush()` and return the combined output in `resp.Output`.

```go
import (
	"strings"
	"testing"

	"github.com/xhd2015/agent-pro/agent/event/print"
)

type Request struct {
	Lines []string
}

type Response struct {
	Output string
}

func Setup(t *testing.T, req *Request) error {
	_ = assertContains
	_ = countSubstring
	return nil
}

func Run(t *testing.T, req *Request) (*Response, error) {
	var state print.FormatState
	var buf strings.Builder
	for _, line := range req.Lines {
		header, body, isMsg := state.FormatLine(line)
		if header == "" && body == "" && !isMsg {
			continue
		}
		if isMsg {
			if header != "" {
				buf.WriteString(header)
				buf.WriteByte('\n')
			}
			buf.WriteString(body)
		} else {
			if header != "" {
				buf.WriteString(header)
				buf.WriteByte('\n')
			}
			if body != "" {
				buf.WriteString(body)
			}
		}
	}
	state.Flush()
	return &Response{Output: buf.String()}, nil
}

func assertContains(t *testing.T, got string, want string) {
	t.Helper()
	if !strings.Contains(got, want) {
		t.Fatalf("missing %q in:\n%s", want, got)
	}
}

func countSubstring(t *testing.T, got string, sub string) int {
	t.Helper()
	return strings.Count(got, sub)
}
```