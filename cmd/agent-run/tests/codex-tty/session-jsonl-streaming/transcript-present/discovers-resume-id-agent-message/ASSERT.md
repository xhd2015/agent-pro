## Expected Output

```text
<contains>
JSONL_DISCOVERED_AGENT_MESSAGE
</contains>
```

## Expected

- Exit code 0.
- Stdout contains the assistant message from `event_msg.agent_message.message`.
- The message is not dependent on cleaned terminal scrollback.

## Side Effects

- `AGENT_RUN_HOME/sessions/codex-tty/.../events.jsonl` records the streamed assistant message.

## Errors

- No error is expected.

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
	assert.Output(t, resp.Stdout, `` +
`<contains>
JSONL_DISCOVERED_AGENT_MESSAGE
</contains>`)
	_, lines := findCodexTTYEventsJSONL(t, req.Home)
	if !eventsContainSubstring(t, lines, codexAgentMessageText) {
		t.Fatalf("expected streamed JSONL assistant message in events.jsonl:\n%s", strings.Join(lines, "\n"))
	}
}
```
