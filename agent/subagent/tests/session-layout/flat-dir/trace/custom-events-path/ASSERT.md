## Expected

- Initial `subagent.Run` succeeds and writes events to `CustomEventsPath`.
- `runTrace` with same `SessionLayout` succeeds.
- Trace stdout contains `Events:` with a positive line count (e.g. `Events:  1 lines`).
- Trace stdout includes the session id and `Done` footer (formatted trace completed).

## Side Effects

- Trace reads external path, not default `Dir/events.jsonl`.

## Errors

- None.

## Exit Code

N/A

```go
import (
	"strings"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	traceOut, traceErr := runTrace(t, req)
	if traceErr != nil {
		t.Fatalf("runTrace failed: %v", traceErr)
	}
	if !strings.Contains(traceOut, "Events:") {
		t.Fatalf("trace output missing Events header:\n%s", traceOut)
	}
	if strings.Contains(traceOut, "Events:  0 lines") {
		t.Fatalf("trace read zero events from custom path:\n%s", traceOut)
	}
	if !strings.Contains(traceOut, req.SessionID) {
		t.Fatalf("trace output missing session id:\n%s", traceOut)
	}
	if !strings.Contains(traceOut, "Done") {
		t.Fatalf("trace output missing Done footer:\n%s", traceOut)
	}
}```
