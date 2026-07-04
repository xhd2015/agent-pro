# Scenario

**Feature**: watch pipe capture shows one screen state, not stacked duplicate redraws

```
# grok-style TUI redraws screen 6 times with true-color incremental updates
harness -> detached fake grok TUI -> watch pipe -> single layout, no smeared duplicates
```

Regression from `renderObserverFrame`: each redraw is converted to plain text and
**appended**, producing duplicate `Grok Build Beta` lines and long smeared braille runs
instead of the latest screen state. Pipe capture must reflect one coherent screen.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Phase = "watch-grok-tui-single-screen-state"
	req.WatchProbe = "3s"
	return nil
}
```