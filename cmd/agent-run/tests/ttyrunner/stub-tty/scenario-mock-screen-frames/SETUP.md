# Scenario

**Feature**: scenario screen_frames mutate scrollback over time

```
screen_frames timed updates -> final frame with prompt
```

## Steps

1. Configure leaf-specific `req` fields.
2. `Run` executes the scenario.

```go
import (
	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Action = "scenario-mock-screen-frames"
	return nil
}
```
