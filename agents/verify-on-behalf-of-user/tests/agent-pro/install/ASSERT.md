## Expected

- Exit code 0.
- `scripts/enter-sandbox.sh` exists under installed skill dir.
- `templates/transcript.md` exists under installed skill dir.
- `transcript/TOPIC.md` exists under installed skill dir.
- `scenario/TOPIC.md` exists under installed skill dir.
- Installed `templates/transcript.md` mentions Depth and browser-agent.

## Exit Code

0

```go
import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	assertExitCode(t, resp, 0)
	var target string
	for _, e := range req.Env {
		if strings.HasPrefix(e, "VERIFY_INSTALL_TARGET=") {
			target = strings.TrimPrefix(e, "VERIFY_INSTALL_TARGET=")
			break
		}
	}
	if target == "" {
		t.Fatal("VERIFY_INSTALL_TARGET not set")
	}
	script := filepath.Join(target, "scripts", "enter-sandbox.sh")
	if _, err := os.Stat(script); err != nil {
		t.Fatalf("missing %s: %v\nstdout:\n%s\nstderr:\n%s", script, err, resp.Stdout, resp.Stderr)
	}
	template := filepath.Join(target, "templates", "transcript.md")
	if _, err := os.Stat(template); err != nil {
		t.Fatalf("missing %s: %v", template, err)
	}
	data, err := os.ReadFile(template)
	if err != nil {
		t.Fatalf("read template: %v", err)
	}
	body := string(data)
	if !strings.Contains(body, "Depth") {
		t.Fatalf("installed transcript template missing Depth field:\n%s", body)
	}
	if !strings.Contains(body, "browser-agent") {
		t.Fatalf("installed transcript template missing browser-agent:\n%s", body)
	}
	topic := filepath.Join(target, "transcript", "TOPIC.md")
	if _, err := os.Stat(topic); err != nil {
		t.Fatalf("missing %s: %v", topic, err)
	}
	scenarioTopic := filepath.Join(target, "scenario", "TOPIC.md")
	if _, err := os.Stat(scenarioTopic); err != nil {
		t.Fatalf("missing %s: %v", scenarioTopic, err)
	}
	ttyTopic := filepath.Join(target, "tty", "TOPIC.md")
	if _, err := os.Stat(ttyTopic); err != nil {
		t.Fatalf("missing %s: %v", ttyTopic, err)
	}
}
```
