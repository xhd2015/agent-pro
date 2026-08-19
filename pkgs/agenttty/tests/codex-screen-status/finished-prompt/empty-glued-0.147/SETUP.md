# Scenario

**Feature**: locked Codex 0.147 empty-glued last line classifies idle

```
› Summarize recent commitsgpt-5.6-terra medium · …
  -> DetectScreenStatus
  -> idle
```

Same last-line shape as `input-box` empty-glued; this leaf locks **screen**,
not occupancy.

## Steps

1. Load the locked 0.147 empty-glued fixture.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Fixture = fixtureEmptyGlued
	return nil
}
```
