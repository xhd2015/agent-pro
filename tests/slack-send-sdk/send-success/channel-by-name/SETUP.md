# Scenario

**Feature**: channel by name `#general` resolves and sends

```
slack-send "#general" -> resolve -> send -> OK
```

## Steps

1. Args `["#general"]`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Args = []string{"#general"}
	return nil
}
```