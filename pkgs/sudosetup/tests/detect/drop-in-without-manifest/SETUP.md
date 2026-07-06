# Scenario

**Feature**: detect orphaned sudoers drop-in without local manifest

```
# drop-in exists, manifest missing
Detect -> Installed=false, detail mentions orphaned drop-in
```

## Preconditions

- Sudoers drop-in file present.
- Manifest file absent.

## Steps

1. Seed sudoers line matching current rule.
2. Do not seed manifest.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Action = "drop_in_without_manifest"
	req.SeedSudoersLine = installedSeedLine("testuser", req.Rule.Command, "")
	return nil
}
```