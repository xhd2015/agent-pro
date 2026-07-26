# Scenario

**Feature**: `run -h` and `resume -h` document `--allow-relocate-resume-session-dir`

```
agent-run run -h     -> stdout contains --allow-relocate-resume-session-dir
agent-run resume -h  -> stdout contains --allow-relocate-resume-session-dir
```

## Steps

1. Primary Run: `agent-run run -h`.
2. Assert also invokes `resume -h` via suite helper.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Args = []string{"run", "-h"}
	return nil
}
```
