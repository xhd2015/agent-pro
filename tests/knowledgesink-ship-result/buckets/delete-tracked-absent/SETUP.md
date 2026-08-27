# Scenario

**Feature**: delete bucket accepts a tracked file removed from disk

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
	req.SeedFiles = map[string]string{
		"SINK.md":   "# sink\n",
		"README.md": "old\n",
	}
	req.DeleteAfterSeed = []string{"README.md"}
	body, err := json.Marshal(knowledgesink.ShipResult{
		HasNewKnowledges: boolPtr(true),
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
	req.ExpectOK = true
	return nil
}
```
