# Scenario

**Feature**: `--log-http` rejects paths without `.jsonl` suffix before grok starts

```
llm-mock run --log-http /tmp/http.log grok -> CLI error (grok not started)
```

## Steps

1. Set `LogHTTPPath` to a non-`.jsonl` path.
2. Do not install fake grok — orchestrator must fail before grok invocation.

```go
import (
	"path/filepath"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.LogHTTPPath = filepath.Join(t.TempDir(), "http.log")
	req.FakeGrokCmd = `sh -c 'echo GROK_RAN; exit 0'`
	return nil
}
```