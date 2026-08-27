# Scenario

**Feature**: object `git_commit_files` with update path validates

```go
import (
	"encoding/json"
	"testing"

	"github.com/xhd2015/agent-pro/pkgs/knowledgesink"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	msg, branch := baseMsgBranch()
	req.SeedFiles = map[string]string{"INDEX.md": "# index\n"}
	body, err := json.Marshal(knowledgesink.ShipResult{
		HasNewKnowledges: boolPtr(true),
		GitCommitMsg:  msg,
		GitBranchName: branch,
		GitCommitFiles: knowledgesink.ShipCommitFiles{
			Update: []string{"INDEX.md"},
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
