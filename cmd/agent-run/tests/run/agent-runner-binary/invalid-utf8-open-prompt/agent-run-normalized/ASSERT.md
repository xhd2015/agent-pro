---
label: real-grok, e2e
explanation: agent-run normalizes invalid UTF-8 before llm-mock-run-grok/real grok; 3s wall-clock no crash; snapshot has message.
---

## Expected

- Real `agent-run` runs under native Go **3s wall-clock** limit (`WallClockLimit`).
- **No** `env.rs` panic in launch output.
- Session `utf8-agent-run-ok` snapshot shows open message
  (`SeaTalk local-bot` / `checkout` / `edgment` / bot check text) and no panic.

## Errors

- env.rs panic → **FAIL**
- Snapshot missing message → **FAIL**

## Exit Code

0 when process still running after 3s (wall-clock stop); or real exit if it finished earlier without panic

```go
import (
	"bytes"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil && !isTimeoutErr(err) {
		t.Fatalf("agent-run failed: %v\nstdout:\n%s\nstderr:\n%s",
			err, resp.Stdout, resp.Stderr)
	}
	combined := ""
	if resp != nil {
		combined = resp.Stdout + "\n" + resp.Stderr
	}
	if containsGrokEnvPanic(combined) {
		t.Fatalf("agent-run→llm-mock-run-grok crashed (env.rs) within 3s:\n%s", combined)
	}

	// Poll snapshot for the open message (PTY may still be painting).
	deadline := time.Now().Add(5 * time.Second)
	var snap string
	for {
		snap = runBinCapture(t, req.AgentRun, req, "snapshot", invalidUTF8AgentSession)
		if containsGrokEnvPanic(snap) {
			t.Fatalf("env.rs panic in snapshot:\n%s", snap)
		}
		if strings.Contains(snap, "SeaTalk local-bot") ||
			strings.Contains(snap, "checkout") ||
			strings.Contains(snap, "check why the bot") ||
			strings.Contains(snap, "edgment") {
			return
		}
		if time.Now().After(deadline) {
			break
		}
		time.Sleep(400 * time.Millisecond)
	}
	t.Fatalf("snapshot missing open message after normalize path:\n%s\nlaunch:\n%s",
		snap, combined)
}

func runBinCapture(t *testing.T, bin string, req *Request, args ...string) string {
	t.Helper()
	cmd := exec.Command(bin, args...)
	cmd.Dir = req.TempDir
	cmd.Env = append(os.Environ(), req.Env...)
	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &buf
	_ = cmd.Run()
	return buf.String()
}
```
