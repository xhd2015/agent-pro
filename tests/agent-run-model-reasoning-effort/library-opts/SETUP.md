# Scenario

**Feature**: `AutoSendOrResume` preserves `Opts.Model` / `Opts.ModelReasoningEffort` (no invent)

```
Opts{Model, ModelReasoningEffort} + RunSession hook
  -> AutoSendOrResume (missing session → ModeRun)
  -> hook captures Model / ModelReasoningEffort unchanged
```

## Steps

1. Set mode `library_opts`.
2. Leaf sets Model and/or Effort.
3. Assert captured fields; empty must not become luna/max.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Mode = "library_opts"
	return nil
}
```
