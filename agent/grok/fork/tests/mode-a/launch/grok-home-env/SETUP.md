# Scenario

**Feature**: `GROK_HOME` on the injected env map is forwarded in the follow-up

```
opts.Env includes GROK_HOME=<fixture>
fork.Main([])
  -> follow-up starts with GROK_HOME=… <exe> --session-id <id>
```

## Steps

1. Set `Env` to `GROK_HOME=<req.GrokHome>`.
2. Bare Mode A args.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Env = []string{"GROK_HOME=" + req.GrokHome}
	req.Args = []string{}
	return nil
}
```
