# Scenario

**Feature**: `GET /` serves the dashboard page

```
usage view --port 0
stdout: serving http://127.0.0.1:<port>
GET /  -> 200 text/html, the dashboard with its chart container
```

## Preconditions

- One grok snapshot in the store, so the page has something to chart.
- The server binds any free port: this leaf asserts the URL line and the page, not
  a particular port.

## Steps

1. Plant one grok snapshot.
2. Start `usage view` in the background and probe the page root.

```go
import (
"strings"
"testing"
"time"

"github.com/xhd2015/agent-pro/agent/usage"
"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
WriteViewSnapshot(t, req, usage.Grok, time.Now().Add(-time.Hour),
map[string]float64{"used_percent": 42, "remaining_percent": 58},
Display("42% used · weekly", "58% remaining"), 178, true, "")
req.HTTPPath = "/"
startViewServer(t, req)
return nil
}
```
