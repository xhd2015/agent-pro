# Agent Hub Tests

These doc-style tests verify the agent-hub CLI for producing events, queueing
messages, consuming events, and querying sessions.

```go
import (
    "bytes"
    "context"
    "encoding/json"
    "errors"
    "fmt"
    "os"
    "os/exec"
    "path/filepath"
    "strconv"
    "strings"
    "testing"
    "time"
	"github.com/xhd2015/doctest/session"
)


type Request struct {
    Home         string
    Command      string
    Args         []string
    Stdin        string
    Env          []string
    TempDir      string
    AgentHub     string
    FakeOpencode string
    RepoRoot     string
    Operation    string
}

type Response struct {
    Stdout   string
    Stderr   string
    ExitCode int
    Err      error
}

func Run(t *testing.T, d *session.Doctest, req *Request) (*Response, error) {
    if req.Operation == "full_workflow" {
        return runFullWorkflow(t, req)
    }
    if req.Command == "" {
        req.Command = req.AgentHub
    }
    return execCmd(t, req.Command, req.Args, req.TempDir, req.Env, req.Stdin)
}
```
