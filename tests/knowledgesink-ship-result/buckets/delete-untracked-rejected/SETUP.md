# Scenario

**Feature**: delete bucket rejects a path that was never tracked

```go
import (
	"encoding/json"
	"testing"

	"github.com/xhd2015/agent-pro/pkgs/knowledgesink"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	msg, branch := baseMsgBranch()
	req.HubMode = "git"
	req.SeedFiles = map[string]string{"SINK.md": "#\n"}
	body, err := json.Marshal(knowledgesink.ShipResult{
		HasNewKnowledges: boolPtr(true),
		GitCommitMsg:  msg,
		GitBranchName: branch,
		GitCommitFiles: knowledgesink.ShipCommitFiles{
			Delete: []string{"never-existed.md"},
		},
	})
	if err != nil {
		return err
	}
	req.ResultJSON = body
	req.ExpectOK = false
	req.ExpectErrSubstr = "not tracked"
	return nil
}
```
