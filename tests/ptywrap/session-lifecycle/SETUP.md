# Scenario

**Feature**: session lifecycle after PTY exit and writer disconnect

```
# exit retention
PTY exits -> status exited -> session remains in GET list until DELETE

# writer disconnect / PTY leak
writer WS close 1000 -> must free child process (PTY)
writer WS close 4000 -> remove session + kill child
create-on-connect churn + close 1000 -> no orphan shells
```

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	// Leaf Setup sets Phase for each case.
	return nil
}
```
