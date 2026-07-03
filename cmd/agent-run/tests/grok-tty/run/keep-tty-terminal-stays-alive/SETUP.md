# Scenario

**Feature**: `--keep-tty` keeps the ptywrap server alive while the process is running (blocks indefinitely). Background run keeps the ptywrap HTTP server reachable.

```
agent-run run --agent-runner grok-tty --keep-tty "hi"
  -> response captured, process blocks
  -> ptywrap server ALIVE on listen_addr (while process is alive)
```

## Preconditions

- Fake TUI (`AGENT_RUN_GROK_TTY_COMMAND`) responds quickly.
- Background process started with `--keep-tty` blocks indefinitely.

## Steps

1. Start `agent-run run --agent-runner grok-tty --keep-tty "hi"` in the background.
2. Wait for the session ID in stderr.
3. `Run` reads registry entry and returns it for TCP reachability check.
4. Process is killed on test cleanup.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.GrokTTYCommand = fakeTUIRespondHi()
	req.GrokTTYPrompt = "keep-alive probe"

	args := []string{"run", "--agent-runner", "grok-tty", "--keep-tty", req.GrokTTYPrompt}

	ctx, cancel := context.WithCancel(context.Background())
	cmd := exec.CommandContext(ctx, req.AgentRun, args...)
	cmd.Dir = req.TempDir
	cmd.Env = append(os.Environ(), req.Env...)
	req.BackgroundStderr = &bytes.Buffer{}
	req.BackgroundStdout = &bytes.Buffer{}
	cmd.Stderr = req.BackgroundStderr
	cmd.Stdout = req.BackgroundStdout
	if err := cmd.Start(); err != nil {
		cancel()
		t.Fatalf("start background grok-tty run: %v", err)
	}
	req.BackgroundCmd = cmd
	t.Cleanup(func() {
		cancel()
		if cmd.Process != nil {
			_ = cmd.Process.Kill()
		}
		_ = cmd.Wait()
	})

	req.GrokTTYSessionID = waitForGrokTTYSessionLine(t, req.BackgroundStderr, 30*time.Second)
	if req.GrokTTYSessionID == "" {
		t.Fatal("failed to get session id from stderr within timeout")
	}

	req.Mode = "registry-while-running"
	return nil
}
```
