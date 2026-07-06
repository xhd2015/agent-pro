# Scenario

**Feature**: Remove deletes sudoers drop-in and manifest, flushes sudo cache

```
Manager.Remove -> sudo rm drop-in -> delete manifest -> sudo -k
```

## Steps

1. Set `Request.Operation = "remove"`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Operation = "remove"
	return nil
}
```