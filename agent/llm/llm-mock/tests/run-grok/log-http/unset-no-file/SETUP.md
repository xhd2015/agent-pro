# Scenario

**Feature**: omitting `--log-http` does not write a log-http JSONL file

```
llm-mock run grok (no --log-http)
fake grok curls mock once -> no session-log-http.jsonl under workdir
```

## Steps

1. Do not set `LogHTTPPath` (flag omitted).
2. Fake grok curls mock once.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.LogHTTPPath = ""
	req.ConfigEnv = "file"
	req.FakeGrokCmd = fakeGrokCurlOnce
	req.ConfigJSON = minimalMockConfigJSON(8080, "")
	return nil
}
```