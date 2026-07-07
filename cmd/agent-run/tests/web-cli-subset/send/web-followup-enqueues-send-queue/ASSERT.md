## Expected

- HTTP response is accepted (202/200).
- Queue file exists under `send-queue/<runner>/<terminal_session_id>.jsonl` containing message text.

```go
import (
	"net/http"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.HTTPStatus != http.StatusAccepted && resp.HTTPStatus != http.StatusOK {
		t.Fatalf("follow-up status=%d body=%q", resp.HTTPStatus, resp.HTTPBody)
	}
	path := queueFilePath(req.Home, req.Runner, req.TerminalSessionID)
	if !queueContainsText(t, path, req.FollowUpPrompt) {
		t.Fatalf("send queue missing follow-up text at %s", path)
	}
}
```
