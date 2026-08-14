# Scenario

**Feature**: library status builder presents input-box for tty status / status

```
DetectInputBox token
  -> InputBoxReport
  -> human "input box: empty|occupied|unknown"
  -> JSON input_box token
```

## Preconditions

- `agent-run tty status` prints `input box: …` **after** sendable.
- `agent-run status` terminal layer uses the same `input_box` field.
- CLI last content line is newline-terminated; `InputBoxReport` human line is
  the label+value without requiring the caller’s trailing `\n` to be stored
  twice (assert the line text; CLI must still end the last line with `\n`).

## Steps

1. Mark `req.Family=report`.
2. Leaves inject a classifiable snapshot or `Unreachable`.

## Context

L2 only — no fake-ptywrap / e2e. Existing `cmd/agent-run/tests/tty/status/` is
not extended.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Family = "report"
	return nil
}
```
