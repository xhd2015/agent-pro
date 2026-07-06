# Scenario

**Feature**: cancel after delivery fails

```
default send (delivered) -> send cancel same id -> exit 1
```

## Steps

1. Set `req.Action = "cancel-after-delivered-fails"`.
2. Set `req.SendMessage = "delivered-probe"`.

```go
func Setup(t *testing.T, req *Request) error {
	req.Action = "cancel-after-delivered-fails"
	req.SendMessage = "delivered-probe"
	return nil
}
```