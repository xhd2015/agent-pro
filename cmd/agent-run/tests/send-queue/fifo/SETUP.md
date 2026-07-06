# Scenario

**Feature**: send queue delivers messages FIFO per session

```
enqueue A --no-wait -> enqueue B --no-wait -> blocking send triggers drainer -> A then B
```

## Preconditions

- Idle stub-tty session with input capture observer attached.

## Steps

1. `Setup` sets `req.Operation = "fifo"`.
2. `Run` enqueues two messages with `--no-wait`, then blocking send to drain.
3. `Assert` verifies injection order A before B.

```go
func Setup(t *testing.T, req *Request) error {
	req.Operation = "fifo"
	req.EnableStubTTY = true
	return nil
}
```