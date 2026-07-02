## Expected Output

Stdout contains formatted stream events from live updates.jsonl tail:

```
<contains>
💬 formatted stream events
⚡ RUN
FORMATTED_ASSISTANT_OUT
</contains>
```

## Expected

- Exit code 0.
- Stderr contains `grok-tty: grok session <uuid>` and `grok-tty: grok updates <path>` once
  discovery succeeds (diagnostics accompany live streaming).
- Stdout contains formatted **user** line (`💬` + prompt text).
- Stdout contains formatted **tool** line (`⚡ RUN` from execute tool_call).
- Stdout contains formatted **assistant** text (`FORMATTED_ASSISTANT_OUT`).
- Output is not empty/silent until end-of-run scrollback fallback only.

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
	assertStderrGrokSession(t, resp.Stderr, formattedGrokUUID, "")
	stdout := resp.Stdout
	if strings.TrimSpace(stdout) == "" {
		t.Fatalf("expected formatted streamed stdout, got empty; stderr:\n%s", resp.Stderr)
	}
	assert.Output(t, stdout, `
<contains>
💬 formatted stream events
⚡ RUN
FORMATTED_ASSISTANT_OUT
</contains>`)
}
```