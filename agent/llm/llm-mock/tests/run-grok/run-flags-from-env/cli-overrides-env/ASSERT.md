## Expected

- Exit code 0.
- CLI path `LogEventsPath` (`b.jsonl`) exists and has at least 1 `type:message` AgentEvent.
- Env path `EnvLogEventsOverridePath` (`a.jsonl`) must not exist.

## Side Effects

- Duplicate `--log-events`: CLI value wins; env path is ignored.

## Exit Code

0

```go
import (
	"os"
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertSuccess(t, resp)

	if req.EnvLogEventsOverridePath == "" {
		t.Fatal("EnvLogEventsOverridePath not set in Setup")
	}
	if _, err := os.Stat(req.EnvLogEventsOverridePath); err == nil {
		t.Fatalf("env log-events path %q must not exist (CLI should win)", req.EnvLogEventsOverridePath)
	} else if !os.IsNotExist(err) {
		t.Fatalf("stat env path %q: %v", req.EnvLogEventsOverridePath, err)
	}

	if len(resp.LogEventsLines) < 1 {
		t.Fatalf("CLI log-events file: want >=1 JSONL line, got %d\ncontent:\n%s",
			len(resp.LogEventsLines), resp.LogEventsContent)
	}

	events, parseErr := parseAgentEventMaps(resp.LogEventsLines)
	if parseErr != nil {
		t.Fatal(parseErr)
	}

	var messageTexts []string
	for _, ev := range events {
		if typ, _ := ev["type"].(string); typ == "message" {
			text, _ := ev["text"].(string)
			messageTexts = append(messageTexts, text)
		}
	}
	if len(messageTexts) < 1 {
		t.Fatalf("want >=1 message AgentEvent, got %#v", events)
	}
	assertContains(t, strings.Join(messageTexts, "\n"), "from-config")
}
```