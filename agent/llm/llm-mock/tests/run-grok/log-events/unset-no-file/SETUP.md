# Scenario

**Feature**: omitting `--log-events` does not write a log-events JSONL file

```
llm-mock run grok (no --log-events)
fake grok curls mock once -> no session-log-events.jsonl under workdir
```

## Steps

1. Do not set `LogEventsPath` (flag omitted).
2. Fake grok performs one curl (mock records request internally only).
3. Assert no `*.jsonl` files appear under workdir after run.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.ConfigEnv = "file"
	req.FakeGrokCmd = fakeGrokCurlOnce
	req.ConfigJSON = minimalMockConfigJSON(8080, "")
	req.LogEventsPath = ""
	return nil
}
```