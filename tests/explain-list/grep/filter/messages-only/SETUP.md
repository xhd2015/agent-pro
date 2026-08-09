# Scenario

**Feature**: --grep searches only Q/A message bodies (not runner/model/dirname)

```
# decoy: agent_runner/model/dirname contain "docker"; bodies do not
# real: body contains "docker"
# explain list --grep docker -> only real kept
```

## Preconditions

- Decoy session with docker in meta fields only; real session with docker in body.

## Steps

1. Seed decoy + real.
2. Args: `list --grep docker`.
3. Assert real marker present; decoy marker absent.

## Context

- Locked search fields: `Messages[].Message` only.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Args = []string{"list", "--grep", "docker"}
	req.Sessions = []SessionSeed{
		{
			// dirname + runner + model all contain "docker"; bodies do not.
			DirName:     "2026-07-08-11-00-00-docker-decoy-dddddddd",
			AgentRunner: "docker-runner",
			Model:       "docker-model",
			Messages: []Msg{
				{Role: "user", Message: "hello marker-decoy"},
				{Role: "assistant", Message: "world without the d-word"},
			},
		},
		simpleSession(
			"2026-07-09-11-00-00-realbody-rrrrrrrr",
			"opencode", "deepseek-chat",
			"I use docker daily marker-real",
			"ok",
		),
	}
	return nil
}
```
