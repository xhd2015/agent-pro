# Scenario

**Feature**: `pkgs/agentruncli` registers `--model-reasoning-effort` flag

```
scan pkgs/agentruncli production .go
  -> contains "--model-reasoning-effort"
```

## Steps

1. Mode `source_wire`, target `cli_flag`.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Mode = "source_wire"
	req.SourceWireTarget = "cli_flag"
	return nil
}
```
