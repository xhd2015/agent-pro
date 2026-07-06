# Scenario

**Feature**: second sequential send prints `msg_2`

```
first send (no-wait) -> msg_1; second send -> stdout msg_2\n
```

## Steps

1. Set `req.Action = "second-send-prints-msg-2"`.
2. Set `req.SendMessage = "second-message"`.

```go
func Setup(t *testing.T, req *Request) error {
	req.Action = "second-send-prints-msg-2"
	req.SendMessage = "second-message"
	return nil
}
```