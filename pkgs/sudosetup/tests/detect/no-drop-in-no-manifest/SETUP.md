# Scenario

**Feature**: detect when neither sudoers drop-in nor manifest exists

```
# empty FS: no drop-in, no manifest
Detect -> Installed=false, verdict password required
```

## Preconditions

- No seeded sudoers drop-in or manifest.

## Steps

1. Leave FS seeds empty (defaults).

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Action = "no_drop_in_no_manifest"
	return nil
}
```