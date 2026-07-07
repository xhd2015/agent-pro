## Expected

- Follow-up HTTP accepted.
- Delivered message text appears in terminal within timeout (queue drainer injection).

```go
import (
	"net/http"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.HTTPStatus != http.StatusAccepted && resp.HTTPStatus != http.StatusOK {
		t.Fatalf("follow-up status=%d body=%q", resp.HTTPStatus, resp.HTTPBody)
	}
	deadline := time.Now().Add(20 * time.Second)
	for time.Now().Before(deadline) {
		cmd := exec.Command(req.AgentRun, "tty", "snapshot", req.TerminalSessionID)
		cmd.Env = append(os.Environ(), req.Env...)
		out, snapErr := cmd.Output()
		if snapErr == nil && strings.Contains(string(out), req.FollowUpPrompt) {
			return
		}
		time.Sleep(200 * time.Millisecond)
	}
	t.Fatalf("timeout waiting for delivered follow-up %q in tty snapshot", req.FollowUpPrompt)
}
```
