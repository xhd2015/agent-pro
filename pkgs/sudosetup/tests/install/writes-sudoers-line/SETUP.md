# Scenario

**Feature**: EnsureInstalled renders sudoers line and runs visudo + install

```
# not installed -> visudo -cf temp -> install drop-in
```

## Preconditions

- No prior install.
- Runner accepts visudo and install.

## Steps

1. Leave seeds empty so install path runs.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Action = "writes_sudoers_line"
	return nil
}
```