# Scenario

**Feature**: `pkgs/agentruncli` registers `--prompt-file` flag

```
scan pkgs/agentruncli production .go
  -> contains "--prompt-file"
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
