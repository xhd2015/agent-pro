# Scenario

**Feature**: Ctrl-C detaches watch when stdin is not in raw mode (SIGINT delivery)

```
# detached sleep session; watch on PTY; harness sends SIGINT; watch should detach
harness PTY -> watch -> SIGINT -> watch detaches; registry entry remains
```

Bug: `streamObserver` calls `signal.Ignore(SIGINT)` on the TTY watch path but only
reads `\x03` from stdin after `term.MakeRaw`. When raw mode is unavailable (stdin
not a TTY, or `MakeRaw` fails), a real Ctrl-C delivers SIGINT — which is ignored —
so watch appears stuck and "nothing happens".

The existing `ctrl-c-detaches` leaf only writes `\x03` to the PTY master, which
bypasses the SIGINT path users hit when stdin stays in canonical mode.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Phase = "watch-ctrl-c-detaches-sigint"
	return nil
}
```