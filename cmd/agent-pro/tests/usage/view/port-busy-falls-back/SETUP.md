# Scenario

**Feature**: a busy `--port` falls back to the next free port with a warning

```
<port held by the test>
usage view --port <held> --no-open
stderr: warning: port <held> is busy, using <held+1>
stdout: serving http://127.0.0.1:<held+1>
```

## Preconditions

- The leaf holds a port itself and passes it to `--port`, so the busy case does not
  depend on port 8080 being free or taken on the machine running the tests.
- An explicit `--port` is a request the server cannot honor, which is why it warns;
  the default 8080 fallback stays quiet because the URL line already reports the
  truth.

## Steps

1. Hold a local port and pass it as `--port`.
2. Probe `/api/healthz` on the URL the server actually printed.

```go
import (
"testing"

"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
port, release := HoldPort(t)
t.Cleanup(release)
req.ViewPort = port
req.HTTPPath = "/api/healthz"
startViewServer(t, req)
return nil
}
```
