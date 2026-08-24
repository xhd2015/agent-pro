# Scenario

**Feature**: delete bucket rejects a path that still exists on disk

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
	req.SeedFiles = map[string]string{"README.md": "still here\n"}
	body, err := json.Marshal(knowledgesink.ShipResult{
		GitCommitMsg:  msg,
		GitBranchName: branch,
		GitCommitFiles: knowledgesink.ShipCommitFiles{
			Delete: []string{"README.md"},
		},
	})
	if err != nil {
		return err
	}
	req.ResultJSON = body
	req.ExpectOK = false
	req.ExpectErrSubstr = "still exists"
	return nil
}
```
