# Scenario

**Feature**: first send on a session prints `msg_1`

```
agent-run send session-N "hello" -> stdout msg_1\n -> drainer delivers when writable
```

## Steps

1. Set `req.Action = "first-send-prints-msg-1"`.
2. Set `req.SendMessage = "hello"`.

```go
func Setup(t *testing.T, req *Request) error {
	req.Action = "first-send-prints-msg-1"
	req.SendMessage = "hello"
	return nil
}
```