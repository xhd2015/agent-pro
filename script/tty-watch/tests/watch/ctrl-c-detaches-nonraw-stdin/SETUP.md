# Scenario

**Feature**: Ctrl-C detaches watch when stdout is a TTY but stdin is not

```
# detached sleep; watch with TTY stdout + pipe stdin; harness sends SIGINT
TTY stdout + pipe stdin -> watch -> SIGINT -> watch detaches; session survives
```

Bug: `streamObserver` enters the TTY observer path when stdout is a terminal, but
skips `term.MakeRaw` when stdin is not a terminal. It still calls
`signal.Ignore(SIGINT)`, so Ctrl-C (SIGINT) is swallowed and watch never detaches.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Phase = "watch-ctrl-c-detaches-nonraw-stdin"
	return nil
}
```