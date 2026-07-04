# Scenario

**Feature**: Observer detach pops grok kitty keyboard protocol stack

```
# read attach.go -> observerTTYDetachCleanup must include \x1b[<u after grok \x1b[?u
```

User report after stdin-restore fix: iTerm2 still shows `^[[?0u` on the prompt and typing
`dddd` / `aa` produces kitty garbage (`d0;1:3u`, `a7;1:3u`, `0u9;5:3u`). Grok enables
kitty keyboard protocol with `\x1b[?u` (push). `\x1b[?0u` alone does not pop that stack
in iTerm2; detach cleanup must emit `\x1b[<u`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Phase = "unit-observer-detach-kitty-pop-cleanup"
	return nil
}
```