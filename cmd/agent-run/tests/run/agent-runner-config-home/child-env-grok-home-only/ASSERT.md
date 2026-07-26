---
label: e2e
---

## Expected

- Exit code 0.
- Captured PTY output contains `GROK_HOME=<config-home-path>`.
- Output does **not** contain `AGENT_RUNNER_CONFIG_HOME=`.

## Exit Code

0

```go
import (
	"strings"
	"testing"
	"github.com/xhd2015/doctest/session"
	"os"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	assertSuccess(t, resp)
	probe, err := os.ReadFile(req.EnvProbePath)
	if err != nil {
		t.Fatalf("read env probe %s: %v", req.EnvProbePath, err)
	}
	dump := string(probe)
	want := "GROK_HOME=" + req.AgentRunnerConfigHome
	if !strings.Contains(dump, want) {
		t.Fatalf("env probe missing %q; probe:\n%s", want, dump)
	}
	if strings.Contains(dump, "AGENT_RUNNER_CONFIG_HOME=") {
		t.Fatalf("env probe must not contain AGENT_RUNNER_CONFIG_HOME=; probe:\n%s", dump)
	}
}
```