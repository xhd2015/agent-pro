# Scenario

**Feature**: DM ID with D prefix passes through

```
resolveChannel("D12345ABCDE") -> D12345ABCDE
```

## Steps

1. Args `["D12345ABCDE"]`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Args = []string{"D12345ABCDE"}
	return nil
}
```