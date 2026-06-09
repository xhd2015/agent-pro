package main

import (
	"strings"
	"testing"

	"github.com/xhd2015/agent-pro/agents/test-case-tree-runner/core"
)

func TestCliDispatchUnknownCommand(t *testing.T) {
	tests := []struct {
		args    []string
		wantErr string
	}{
		{[]string{"compile"}, "unknown command"},
		{[]string{"run"}, "unknown command"},
		{[]string{"xyz"}, "unknown command"},
	}

	for _, tt := range tests {
		err := runCli(tt.args)
		if err == nil {
			t.Fatalf("expected error for args %v", tt.args)
		}
		if !strings.Contains(err.Error(), tt.wantErr) {
			t.Fatalf("args %v: expected error containing %q, got %v", tt.args, tt.wantErr, err)
		}
	}
}

func TestCliDispatchHelp(t *testing.T) {
	for _, args := range [][]string{
		{},
		{"-h"},
		{"--help"},
	} {
		err := runCli(args)
		if err != nil {
			t.Fatalf("expected no error for args %v, got %v", args, err)
		}
	}
}

func TestCliDispatchTestRequiresDir(t *testing.T) {
	err := runCli([]string{"test"})
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "test requires") {
		t.Fatalf("expected 'test requires' error, got %v", err)
	}
}

func TestCliDispatchBuildRequiresDir(t *testing.T) {
	err := runCli([]string{"build"})
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "build requires") {
		t.Fatalf("expected 'build requires' error, got %v", err)
	}
}

func TestCliDispatchGenerateCodeRequiresDir(t *testing.T) {
	err := runCli([]string{"generate-code"})
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "generate-code requires") {
		t.Fatalf("expected 'generate-code requires' error, got %v", err)
	}
}

func TestParseTestOptionsBasic(t *testing.T) {
	opts, remainArgs, err := parseTestOptions([]string{"mydir"})
	if err != nil {
		t.Fatalf("parseTestOptions: %v", err)
	}
	if len(remainArgs) != 1 || remainArgs[0] != "mydir" {
		t.Fatalf("expected [mydir], got %v", remainArgs)
	}
	if opts.Verbose {
		t.Fatal("expected verbose=false")
	}
	if opts.RemoveTemp {
		t.Fatal("expected RemoveTemp=false")
	}
}

func TestParseTestOptionsWithFlags(t *testing.T) {
	opts, remainArgs, err := parseTestOptions([]string{"-v", "--rm", "mydir"})
	if err != nil {
		t.Fatalf("parseTestOptions: %v", err)
	}
	if len(remainArgs) != 1 || remainArgs[0] != "mydir" {
		t.Fatalf("expected [mydir], got %v", remainArgs)
	}
	if !opts.Verbose {
		t.Fatal("expected verbose=true")
	}
	if !opts.RemoveTemp {
		t.Fatal("expected RemoveTemp=true")
	}
}

func TestParseTestOptionsCount(t *testing.T) {
	opts, remainArgs, err := parseTestOptions([]string{"-count=5", "mydir"})
	if err != nil {
		t.Fatalf("parseTestOptions: %v", err)
	}
	if len(remainArgs) != 1 || remainArgs[0] != "mydir" {
		t.Fatalf("expected [mydir], got %v", remainArgs)
	}
	if opts.Count != 5 {
		t.Fatalf("expected count=5, got %d", opts.Count)
	}
}

func TestParseBuildOptionsGenDir(t *testing.T) {
	opts, remainArgs, err := parseBuildOptions([]string{"--gen-dir", "/tmp/gen", "mydir"})
	if err != nil {
		t.Fatalf("parseBuildOptions: %v", err)
	}
	if len(remainArgs) != 1 || remainArgs[0] != "mydir" {
		t.Fatalf("expected [mydir], got %v", remainArgs)
	}
	if opts.GenDir != "/tmp/gen" {
		t.Fatalf("expected GenDir=/tmp/gen, got %q", opts.GenDir)
	}
}

func TestTestTreeUsage(t *testing.T) {
	testsDir := t.TempDir()
	writeTreeFile(t, testsDir, "README.md", "# tree")
	writeTreeFile(t, testsDir, "SETUP.md", setupDoc(`
type Request struct{}
type Response struct{}
func Run(t *testing.T, req *Request) (*Response, error) { return &Response{}, nil }
`))
	writeTreeFile(t, testsDir, "leaf/SETUP.md", setupDoc(`
func Setup(t *testing.T, req *Request) error { _ = req; return nil }
`))
	writeTreeFile(t, testsDir, "leaf/ASSERT.md", assertDoc(`
func Assert(t *testing.T, req *Request, resp *Response, err error) {}
`))

	if err := TestTree(testsDir, core.Options{RemoveTemp: true}); err != nil {
		t.Fatalf("test tree: %v", err)
	}
}
