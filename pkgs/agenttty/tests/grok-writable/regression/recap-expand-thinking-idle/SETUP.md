# Scenario

**Bug**: post-turn Recap + `Ctrl+e:expand thinking` footer must stay idle/sendable

```
snapshot: Worked for + Recap + ❯ Build anything + Ctrl+e:expand thinking
  -> CheckWritable: ready=true, state=idle
  -> ClassifyGrokScreen = idle
  -> OpenReady = true
```

Crime scene: live session `01a03d6f-c449-7bd0-9305-d41474fd1940` (`kck grok snapshot`).
Turn finished (`Worked for` / stop hooks); footer phrase `expand thinking` must not
match busy substring `thinking`. Composer text `Build anything` is placeholder chrome.

## Preconditions

- Fixture `grok-after_recap-expand-thinking-idle-01a03d6f.txt` is sanitized snapshot text.

## Steps

1. Set `req.FixtureFile` to the recap expand-thinking idle fixture.

## Context

- Desired idle; RED on current checkGrokWritable / ClassifyGrokScreen until detector
  ignores `expand thinking` footer phrase.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.FixtureFile = fixtureRecapExpandThinkingIdle
	return nil
}
```
