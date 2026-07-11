# Scenario

**Feature**: A2 — two `--no-wait` enqueues deliver FIFO without a blocking trigger send

```
--no-wait A -> --no-wait B -> (CLI done) session drainer -> inject A then B
```

## Steps

1. Leaf uses default `fifo` operation (no extra Setup fields).
2. Harness must not issue a third default/blocking send to "wake" a CLI drainer.

```go
import "time"

func Setup(t *testing.T, req *Request) error {
	req.ExecTimeout = 60 * time.Second
	return nil
}
```
