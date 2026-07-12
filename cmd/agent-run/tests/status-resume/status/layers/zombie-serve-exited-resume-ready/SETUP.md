# Scenario

**Bug**: after agent `/exit`, keep-alive serve stays reachable → status must still report `exited: true` and resume ready

```
meta running + runner_session_id
  + registry: alive PID (serve)
  + fake ptywrap reachable
  + scrollback: "grok --resume <id>" + "[Terminal exited]" (sendable no)
  -> agent-run status test-zombie-s1
  -> process: alive (serve)
  -> terminal: reachable, sendable: no
  -> runner: bound, exited: true
  -> resume: ready: yes
```

## Steps

1. Start fake ptywrap with post-exit scrollback (exit markers, no idle prompt).
2. Write live registry (alive PID = serve keep-alive).
3. Seed bound meta (`runner_session_id` set, status running).
4. Run human `status`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.SessionID = "test-zombie-s1"
	req.RunnerSessionID = "550e8400-e29b-41d4-a716-446655440222"
	req.TerminalSessionID = "term-zombie-1"
	req.InitialPrompt = "zombie after exit"
	seedZombieServeAfterExit(t, req)
	req.Args = []string{"status", req.SessionID}
	return nil
}
```
