package faketools

import (
	"fmt"
	"math/rand"
)

type FileChange struct {
	Path string
	Kind string
}

type Result struct {
	Command string
	Output  string
	Changes []FileChange
}

type fakeEntry struct {
	cmd  string
	text string
}

var toolOutputs = []fakeEntry{
	{"ls -la", "src/\n  main.go\n  utils.go\nREADME.md\n"},
	{"git status", "On branch main\nnothing to commit, working tree clean\n"},
	{"git diff", ""},
	{"cat README.md", "# Project\n\nThis is a sample project.\nSee docs/ for /tmp/details.\n"},
	{"find . -name \"*.go\"", "./cmd/main.go\n./pkgs/foo/foo.go\n"},
	{"go test ./...", "ok \t github.com/xhd2015/agent-pro/... \t 0.123s\n"},
	{"go build ./...", ""},
}

var fileContents = []fakeEntry{
	{"cat config.json", "{\n  \"version\": \"1.0\",\n  \"include\": [\"/tmp/shared\", \"/tmp/plugins\"]\n}\n"},
	{"cat Makefile", "build:\n\tgo build -o /tmp/output ./cmd/...\n\ntest:\n\tgo test ./...\n"},
	{"cat .env", "DATABASE_URL=postgres://localhost/db\nAPI_KEY=sk-xxx\n"},
	{"cat TODO.md", "# TODO\n\n- fix bug in /tmp/auth.go\n- add /tmp/feature.go\n- search for deprecated\n"},
}

var searchResults = []string{
	"src/main.go:10: import \"github.com/xhd2015/agent-pro\"\nsrc/main.go:42: TODO: refactor /tmp/legacy.go\n",
	"README.md:5: See /tmp/docs/guide.md for setup.\nREADME.md:20: Run `go test ./tmp/...`\n",
	"config.toml:3: path = \"/tmp/data\"\nconfig.toml:8: include = \"/tmp/extra\"\n",
	"",
}

var defaultCommands = []string{
	"ls -la",
	"git status",
	"git diff",
	"cat README.md",
	"find . -name \"*.go\"",
	"go test ./...",
	"go build ./...",
}

func ExecToolCall(rng *rand.Rand, cmd string) Result {
	e := toolOutputs[rng.Intn(len(toolOutputs))]
	return Result{
		Command: cmd,
		Output:  e.text,
	}
}

func ExecFileRead(rng *rand.Rand, path string) Result {
	e := fileContents[rng.Intn(len(fileContents))]
	return Result{
		Command: fmt.Sprintf("cat %s", path),
		Output:  e.text,
	}
}

func ExecFileWrite(rng *rand.Rand, path string) Result {
	kind := "add"
	if rng.Intn(2) == 0 {
		kind = "modify"
	}
	return Result{
		Changes: []FileChange{{Path: path, Kind: kind}},
	}
}

func ExecSearch(rng *rand.Rand, query string) Result {
	s := searchResults[rng.Intn(len(searchResults))]
	return Result{
		Command: fmt.Sprintf("grep -rn %s .", query),
		Output:  s,
	}
}

func ExecRandom(rng *rand.Rand) Result {
	cmd := defaultCommands[rng.Intn(len(defaultCommands))]
	return ExecToolCall(rng, cmd)
}
