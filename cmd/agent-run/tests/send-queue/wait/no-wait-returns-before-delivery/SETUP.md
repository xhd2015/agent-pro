# Scenario

**Feature**: `--no-wait` returns before message delivery on busy session

```
busy terminal + --no-wait -> <1s, id printed, message not yet injected
```

## Steps

1. Set `req.Action = "no-wait-returns-before-delivery"`.
2. Set `req.SendMessage = "no-wait-probe"`.

```go
func Setup(t *testing.T, req *Request) error {
	req.Action = "no-wait-returns-before-delivery"
	req.SendMessage = "no-wait-probe"
	return nil
}
```