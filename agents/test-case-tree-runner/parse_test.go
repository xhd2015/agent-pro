package main

import (
	"strings"
	"testing"

	"github.com/xhd2015/agent-pro/agents/test-case-tree-runner/core"
)

func TestExtractFinalGoBlockIgnoresSectionNames(t *testing.T) {
	block, err := core.ExtractFinalGoBlock("SETUP.md", "## Anything\n\ntext\n\n```go\nfunc Setup(t *testing.T, req *Request) error { return nil }\n```\n")
	if err != nil {
		t.Fatalf("extract final go block: %v", err)
	}
	if !strings.Contains(block.Code, "func Setup") {
		t.Fatalf("expected setup code, got %q", block.Code)
	}
}

func TestSetupRejectsMultipleGoBlocks(t *testing.T) {
	_, err := core.ParseSetupDocument("SETUP.md", "```go\nvar x = 1\n```\n\n```go\nfunc Setup(t *testing.T, req *Request) error { return nil }\n```\n")
	if err == nil {
		t.Fatal("expected error for multiple go blocks")
	}
	if !strings.Contains(err.Error(), "multiple") {
		t.Fatalf("expected multiple go blocks error, got %v", err)
	}
}

func TestSetupRejectsNonFinalGoBlock(t *testing.T) {
	_, err := core.ParseSetupDocument("SETUP.md", "```go\nfunc Setup(t *testing.T, req *Request) error { return nil }\n```\n\nmore markdown\n")
	if err == nil {
		t.Fatal("expected error for non-final go block")
	}
	if !strings.Contains(err.Error(), "final") {
		t.Fatalf("expected final go block error, got %v", err)
	}
}

func TestAssertRequiresSingleFinalGoBlock(t *testing.T) {
	_, err := core.ParseAssertDocument("ASSERT.md", "## Expected\n- no executable assertion\n")
	if err == nil {
		t.Fatal("expected error for missing assert go block")
	}
	if !strings.Contains(err.Error(), "go block") {
		t.Fatalf("expected go block error, got %v", err)
	}
}

func TestSetupFunctionSignature(t *testing.T) {
	_, err := core.ParseSetupDocument("SETUP.md", setupDoc(`
type Request struct{}
type Response struct{}
func Setup(t *testing.T) {}
`))
	if err == nil {
		t.Fatal("expected invalid setup signature error")
	}
	if !strings.Contains(err.Error(), "Setup") || !strings.Contains(err.Error(), "*Request") {
		t.Fatalf("expected setup signature error, got %v", err)
	}
}

func TestRunFunctionSignature(t *testing.T) {
	_, err := core.ParseSetupDocument("SETUP.md", setupDoc(`
type Request struct{}
type Response struct{}
func Run(t *testing.T, req *Request) error { return nil }
`))
	if err == nil {
		t.Fatal("expected invalid run signature error")
	}
	if !strings.Contains(err.Error(), "Run") || !strings.Contains(err.Error(), "*Response") {
		t.Fatalf("expected run signature error, got %v", err)
	}
}

func TestAssertFunctionSignature(t *testing.T) {
	_, err := core.ParseAssertDocument("ASSERT.md", assertDoc(`
func Assert(t *testing.T, req *Request, err error) {}
`))
	if err == nil {
		t.Fatal("expected invalid assert signature error")
	}
	if !strings.Contains(err.Error(), "Assert") || !strings.Contains(err.Error(), "*Response") {
		t.Fatalf("expected assert signature error, got %v", err)
	}
}

func TestInvalidGoReportsSourceMarkdown(t *testing.T) {
	_, err := core.ParseSetupDocument("branch/SETUP.md", setupDoc(`
type Request struct{
`))
	if err == nil {
		t.Fatal("expected invalid go error")
	}
	if !strings.Contains(err.Error(), "branch/SETUP.md") {
		t.Fatalf("expected source markdown path in error, got %v", err)
	}
}
