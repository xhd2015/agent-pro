# Scenario

**Feature**: help prints usage without loading config or sending

```
caller -> slack-send -h|--help -> printHelp -> stdout usage block -> exit 0
```

## Preconditions

- Help is handled before config discovery.

## Steps

1. No config fixture; `WriteGoMod` false.
2. Leaf sets `-h` or `--help` in `req.Args`.

## Context

- Help text includes hardcoded defaults `C0ALE44K5J6` and `#general`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.WriteGoMod = false
	req.ConfigFixture = ""
	req.ConfigInline = ""
	return nil
}
```