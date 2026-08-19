# Scenario

**Feature**: llm-mock-run-codex mock-model `›` chrome classifies idle

```
› reply with exactly: pong
› Use /skills …mock-model default · …
  -> DetectScreenStatus
  -> idle
```

Scene chrome has no `CODEX_TTY_BANNER`. Occupancy may stay occupied
(`default ·` is not ` medium · ` glue) — this leaf only locks screen.

## Steps

1. Load the scene mock-model snapshot fixture.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Fixture = fixtureMockModelDefault
	return nil
}
```
