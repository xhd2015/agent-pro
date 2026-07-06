# gen-commit-msg Tests

Doc-style tests for `github.com/xhd2015/agent-pro/agent/commit_msg`, covering
AI commit message generation with `fake-opencode` and the `--commit` git commit path.

# DSN (Domain Specific Notion)

`gen-commit-msg` stages a diff, calls an opencode-compatible agent runner to
produce a JSON commit message, optionally runs `git commit`. Tests use
`fake-opencode` via `--agent-runner-binary` and `FAKE_OPENCODE_MOCK_CONFIG` to
emit deterministic `step_start` / `text` / `step_finish` events. Race tests
spawn a background `git status` loop during the agent phase to reproduce
`index.lock` contention at commit time. The `--no-verify` flag forwards to
`git commit --no-verify` when used with `--commit`; alone it must fail before
the agent runs.

## Version

0.0.2

## Decision Tree

```
agent/commit_msg/tests/
├── DOCTEST.md
├── SETUP.md
├── auto-unstage/
│   └── subdir-repo-root-paths/ nested --dir + repo-root staged paths → auto unstage succeeds
├── commit-with-fake-opencode/
│   └── succeeds/              fake-opencode returns JSON → message printed, no --commit
├── commit-race/
│   ├── background-git-loop/   agent spawns background git loop → --commit must succeed
│   └── worktree-concurrent-git/
│                               git worktree + background git loop → --commit must succeed
└── no-verify/
    ├── requires-commit/       --no-verify without --commit → early error, no agent
    └── passes-to-git/         failing pre-commit hook + --commit --no-verify → commit succeeds
```

## Test Index

| # | Leaf | Description |
|---|------|-------------|
| 1 | `auto-unstage/subdir-repo-root-paths` | Nested `--dir` with repo-root staged paths must not fail auto unstage |
| 2 | `commit-with-fake-opencode/succeeds` | fake-opencode mock returns commit JSON; generation succeeds |
| 3 | `commit-race/background-git-loop` | Background git status loop during agent run must not break `--commit` |
| 4 | `commit-race/worktree-concurrent-git` | Same race reproduced from a linked git worktree |
| 5 | `no-verify/requires-commit` | `--no-verify` without `--commit` errors before agent |
| 6 | `no-verify/passes-to-git` | `--commit --no-verify` skips failing pre-commit hook |

## How to Run

```sh
doctest vet ./agent/commit_msg/tests
doctest test -v ./agent/commit_msg/tests
doctest test -v ./agent/commit_msg/tests/commit-race/background-git-loop
doctest test -v ./agent/commit_msg/tests/no-verify/requires-commit
doctest test -v ./agent/commit_msg/tests/no-verify/passes-to-git
```

```go
import "testing"

type Request struct {
	RepoRoot          string
	TempDir           string
	GitDir            string
	WorktreeDir       string
	FakeOpencode      string
	MockConfigPath    string
	Model             string
	Commit            bool
	NoVerify          bool
	Operation         string
	AgentRunner       string
	AgentRunnerBinary string
}

type Response struct {
	Stdout   string
	Stderr   string
	Message  string
	Err      error
	ExitCode int
}

func Run(t *testing.T, req *Request) (*Response, error) {
	return captureRunGenCommitMsg(t, req)
}
```