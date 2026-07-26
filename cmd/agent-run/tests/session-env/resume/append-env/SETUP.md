# Scenario

**Feature**: resume `-e` appends to stored env and persists

```
seed meta.env=[FOO=bar]
  -> resume -e NEW=1 <id> "followup"
  -> child FOO=bar and NEW=1
  -> meta.env grows with NEW=1
```

## Steps

1. Seed meta with `env: ["FOO=bar"]`.
2. Resume with extra `-e NEW=1`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	prepareEnvLoggingResume(t, req)

	req.SessionID = "sess-env-append-env"
	req.SeedEnv = []string{"FOO=bar"}
	seedBoundExitedMeta(t, req)

	req.Prompt = "resume append env"
	req.Args = []string{
		"resume",
		"-e", "NEW=1",
		"--agent-runner-binary", req.AgentRunnerBinary,
		req.SessionID,
		req.Prompt,
	}
	return nil
}
```
