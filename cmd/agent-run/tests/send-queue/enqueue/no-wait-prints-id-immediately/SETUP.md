# Scenario

**Feature**: `--no-wait` prints message id and returns quickly

```
busy terminal + --no-wait -> stdout msg_N\n within ~1s
```

## Steps

1. Set `req.Action = "no-wait-prints-id-immediately"`.
2. Set `req.SendMessage = "no-wait-probe"`.

```go
func Setup(t *testing.T, req *Request) error {
	req.Action = "no-wait-prints-id-immediately"
	req.SendMessage = "no-wait-probe"
	return nil
}
```