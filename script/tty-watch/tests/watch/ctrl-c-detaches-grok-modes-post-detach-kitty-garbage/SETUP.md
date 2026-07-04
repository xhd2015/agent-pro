# Scenario

**Feature**: Post-detach shell does not receive kitty key/mouse garbage

```
# grok modes -> watch -> iTerm kitty Ctrl-C -> inject iTerm kitty key bytes -> no garbage
```

User typing after watch exit in iTerm2 showed `^[[?0u` prompt smear and kitty fragments
(`0u9;5:3u`, `d0;1:3u`, `a7;1:3u`) when typing `dddd` / `aa`. Harness injects the kitty
keyboard protocol bytes iTerm2 delivers while kitty mode remains active after incomplete
cleanup (`\x1b[?0u` without `\x1b[<u` pop).

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Phase = "watch-ctrl-c-detaches-grok-modes-post-detach-kitty-garbage"
	return nil
}
```