# Scenario

**Feature**: ttyrunner.IsTTYRunner distinguishes TTY vs non-TTY runners

```
IsTTYRunner(grok-tty) -> true; IsTTYRunner(fake-codex) -> false
```

## Steps

1. Configure leaf-specific `req` fields.
2. `Run` executes the scenario.

```go
func Setup(t *testing.T, req *Request) error {
	req.Action = "is-tty-runner"
	req.RunnerID = "grok-tty"
	return nil
}
```
