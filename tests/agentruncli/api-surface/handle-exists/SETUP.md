# Scenario

**Feature**: Handle entrypoint exists and is callable

```
# touch public entry
agentruncli.Handle(["--help"])
  -> compile + callable (err may be nil on help)
```

## Preconditions

- No store seed, binary build, or network required.

## Steps

1. Mode `api_surface`.
2. Args `--help` so a successful extract returns nil without side-effect commands.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.Mode = "api_surface"
	req.Args = []string{"--help"}
	return nil
}
```
