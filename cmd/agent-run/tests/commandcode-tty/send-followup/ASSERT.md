---
label: e2e
---

## Expected

- `send` prints `msg_1` and exits 0.
- Snapshot after send contains both "Hello" and "Hello 2".

## Exit Code

0

```go
import (
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/xhd2015/doctest/assert"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	sessionID := req.SessionID
	childEnv := append(os.Environ(), req.Env...)

	// Step 1: open session with --session-id
	openArgs := []string{
		"run",
		"--agent-runner", "commandcode-tty",
		"--agent-runner-binary", req.MockBinary,
		"--session-id", sessionID,
		"--open", "Hello",
	}
	t.Logf("session id: %s", sessionID)
	openCmd := exec.Command(req.AgentRun, openArgs...)
	openCmd.Env = childEnv
	openOut, err := openCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("run --open failed: %v\n%s", err, string(openOut))
	}
	t.Logf("open output: %s", string(openOut))

	time.Sleep(2 * time.Second)

	// Step 2: send
	sendArgs := []string{"send", "--max-wait", "15s", sessionID, "Hello 2"}
	sendCmd := exec.Command(req.AgentRun, sendArgs...)
	sendCmd.Env = childEnv
	sendOut, err := sendCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("send failed: %v\n%s", err, string(sendOut))
	}
	sendResult := strings.TrimSpace(string(sendOut))
	if !strings.HasPrefix(sendResult, "msg_") {
		t.Fatalf("expected msg_N, got: %s", sendResult)
	}
	t.Logf("send result: %s", sendResult)

	time.Sleep(2 * time.Second)

	// Step 3: snapshot
	snapshot := execSnapshot(t, req.AgentRun, sessionID, req.Env...)
	if !strings.Contains(snapshot, "Hello") {
		t.Fatalf("snapshot missing 'Hello' for %s:\n%s", sessionID, snapshot)
	}
	if !strings.Contains(snapshot, "Hello 2") {
		t.Fatalf("snapshot missing 'Hello 2' for %s:\n%s", sessionID, snapshot)
	}

	resp.Snapshot = snapshot
	resp.SessionID = sessionID
	resp.ExitCode = 0
}
```
