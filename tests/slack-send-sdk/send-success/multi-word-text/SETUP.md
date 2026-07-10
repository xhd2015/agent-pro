# Scenario

**Feature**: multi-word text joined from remaining args

```
slack-send "#general" "Hello from Go debug script" -> text joined with spaces -> OK
```

## Steps

1. Args `["#general", "Hello from Go debug script"]`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Args = []string{"#general", "Hello from Go debug script"}
	return nil
}
```