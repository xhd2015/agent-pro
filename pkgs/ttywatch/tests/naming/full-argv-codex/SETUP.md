# Scenario

**Feature**: codex argv with flags produces joined slug and wrapped serve token

```
argv [codex run --model gpt-5.5 medium]
  -> SlugifyCommandLine -> codex_run_model_gpt-5.5_medium
  -> ServeSubcommand -> __serve_codex_run_model_gpt-5.5_medium__
```

## Steps

1. Set `req.Argv` to a typical codex invocation with `--model` flag.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Argv = []string{"codex", "run", "--model", "gpt-5.5", "medium"}
	return nil
}
```