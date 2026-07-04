# Scenario

**Feature**: watch streams raw PTY output to stdout

```
# detached loop echoes marker
harness -> detached echo loop -> watch -> WATCH_MARKER on stdout
```

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Phase = "watch-stream"
	req.WatchProbe = "4s"
	return nil
}
```