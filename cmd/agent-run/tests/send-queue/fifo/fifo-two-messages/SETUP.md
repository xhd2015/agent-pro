# Scenario

**Feature**: two no-wait enqueues deliver in FIFO order

```
--no-wait A -> --no-wait B -> blocking trigger -> inject A then B
```

## Steps

1. Leaf uses default `fifo` operation (no extra Setup fields).

```go
import "time"

func Setup(t *testing.T, req *Request) error {
	req.ExecTimeout = 60 * time.Second
	return nil
}
```