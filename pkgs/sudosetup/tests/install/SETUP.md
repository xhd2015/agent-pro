# Scenario

**Feature**: EnsureInstalled writes sudoers drop-in and manifest when needed

```
# skip if installed; else temp -> visudo -cf -> install -> manifest
Manager.EnsureInstalled -> [visudo + install] -> manifest JSON
```

## Steps

1. Set `Request.Operation = "install"`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Operation = "install"
	return nil
}
```