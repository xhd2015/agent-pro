# Scenario

**Feature**: public package surface for agentruncli

```
# import library package
import github.com/xhd2015/agent-pro/pkgs/agentruncli
  -> Handle(args []string) error exists
  -> package name is agentruncli (not main)
```

## Preconditions

- Package path `github.com/xhd2015/agent-pro/pkgs/agentruncli` must compile as a library.
- Public symbol `Handle` must exist (see root DOCTEST planned API).

## Steps

1. Set harness mode for api-surface leaves (`api_surface` or leaf-specific).
2. `Run` touches `Handle` or scans package clause.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	// Default grouping mode; not-package-main leaf overrides.
	if req.Mode == "" {
		req.Mode = "api_surface"
	}
	return nil
}
```
