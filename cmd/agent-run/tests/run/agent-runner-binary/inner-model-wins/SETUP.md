# Scenario

**Feature**: inner `--model` in binary spec wins over CLI `--model`

```
agent-run run --model outer --agent-runner-binary "script --model inner"
  -> spawned argv uses inner, not outer
```

## Steps

1. Write model-probe runner with `--model inner` baked into the binary spec string.
2. Pass CLI `--model outer` on `agent-run run`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	binDir := filepath.Join(req.TempDir, "fake-bin")
	if err := os.MkdirAll(binDir, 0755); err != nil {
		return err
	}
	req.ArgvProbePath = filepath.Join(req.TempDir, "argv-probe.log")
	script := writeModelProbeRunner(t, binDir, "inner", req.ArgvProbePath)
	req.AgentRunnerBinary = script + " --model inner"
	req.Prompt = "model precedence"
	req.Args = append(req.Args,
		"--model", "outer",
		"--agent-runner-binary", req.AgentRunnerBinary,
		req.Prompt,
	)
	return nil
}
```