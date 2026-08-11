# Scenario

**Feature**: WireOnTTYStarted + AutoSendOrResume true-TTY path publishes bus event

```
# URL set
WireOnTTYStarted(URL) -> Opts.OnTTYStarted
AutoSendOrResume(ModeRun) -> library fires hook once
  -> PublishCount=1; type agent.tty.started source agent-run

# URL empty
WireOnTTYStarted("") + ModeRun -> PublishCount=0
```

## Preconditions

- Not ForceNew-only: library `OnTTYStarted` fires on ModeRun establish.
- `agentruncli.WireOnTTYStarted` returns callback for `agentrunapi.Opts.OnTTYStarted`.
- Injectable PublishHook; no real HTTP / iTerm / agent binary.
- Leaves set EventBusURL empty or non-empty.

## Steps

1. Grouping sets `Op=library-wire`, inject publisher.
2. Leaf sets URL / session identity.
3. Run wires OnTTYStarted and ModeRun AutoSendOrResume.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.Op = opLibraryWire
	req.UseInjectPublisher = true
	if req.Capture == nil {
		req.Capture = &HTTPCapture{}
	}
	if req.SessionID == "" {
		req.SessionID = "sess-library-wire-1"
	}
	if req.Runner == "" {
		req.Runner = "grok-tty"
	}
	if req.Workspace == "" {
		req.Workspace = "/tmp/ws-library-wire"
	}
	return nil
}
```
