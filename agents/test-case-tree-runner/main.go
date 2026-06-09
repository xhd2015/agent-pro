package main

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/xhd2015/less-flags"
	"github.com/xhd2015/agent-pro/agents/test-case-tree-runner/generate"
)

const usage = `Usage: test-case-tree-runner <command> [options]

Commands:
  run [-v] [--rm] [-count=N] <dir>
                             Run executable Go snippets from a test-case-tree directory.
  compile [-v] [--rm] [--gen-dir DIR] [-count=N] <dir>
                             Validate generated snippets compile without executing behavior.
  generate-code [--dry-run] <dir>
                             Fill in missing Go code blocks using the test-case-tree-design-expert agent.
`

type GoBlock struct {
	SourcePath string
	Code       string

	Imports   []string
	TypeDecls []string
	VarDecls  []string
	Consts    []string
	Helpers   []FuncSnippet
	Setup     *FuncSnippet
	Run       *FuncSnippet
	Assert    *FuncSnippet

	Types map[string]bool
}

type FuncSnippet struct {
	Name    string
	Params  string
	Results string
	Body    string
}

type SetupDocument struct {
	Path    string
	GoBlock *GoBlock
}

type AssertDocument struct {
	Path    string
	GoBlock GoBlock
}

type TreeCase struct {
	Name       string
	Path       string
	SetupFiles []SetupDocument
	AssertFile AssertDocument
}

type CompileOptions struct {
	GenDir     string
	Verbose    bool
	Stderr     io.Writer
	RemoveTemp bool
	Count      int
}

type ValidationError struct {
	Path string
	Msg  string
}

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) == 0 || args[0] == "-h" || args[0] == "--help" {
		fmt.Print(usage)
		return nil
	}
	switch args[0] {
	case "run":
		opts, cmdArgs, err := parseRunOptions(args[1:])
		if err != nil {
			return err
		}
		if len(cmdArgs) != 1 {
			return fmt.Errorf("run requires <dir>")
		}
		return RunTree(cmdArgs[0], opts)
	case "compile":
		opts, cmdArgs, err := parseCompileOptions(args[1:])
		if err != nil {
			return err
		}
		if len(cmdArgs) != 1 {
			return fmt.Errorf("compile requires <dir>")
		}
		return CompileTreeWithOptions(cmdArgs[0], opts)
	case "generate-code":
		var opts generate.GenerateOptions
		cmdArgs := args[1:]
		for len(cmdArgs) > 0 && strings.HasPrefix(cmdArgs[0], "-") {
			switch cmdArgs[0] {
			case "--dry-run":
				opts.DryRun = true
				cmdArgs = cmdArgs[1:]
			default:
				return fmt.Errorf("unknown generate-code option: %s", cmdArgs[0])
			}
		}
		if len(cmdArgs) != 1 {
			return fmt.Errorf("generate-code requires <dir>")
		}
		return generate.Run(cmdArgs[0], opts)
	default:
		return fmt.Errorf("unknown command: %s", args[0])
	}
}

func parseRunOptions(args []string) (CompileOptions, []string, error) {
	opts := CompileOptions{Stderr: os.Stderr}
	remainArgs, err := lessflags.Bool("-v,--verbose", &opts.Verbose).
		Bool("--rm", &opts.RemoveTemp).
		Int("-count", &opts.Count).
		Parse(args)
	if err != nil {
		return opts, nil, err
	}
	return opts, remainArgs, nil
}

func parseCompileOptions(args []string) (CompileOptions, []string, error) {
	opts := CompileOptions{Stderr: os.Stderr}
	remainArgs, err := lessflags.Bool("-v,--verbose", &opts.Verbose).
		Bool("--rm", &opts.RemoveTemp).
		String("--gen-dir", &opts.GenDir).
		Int("-count", &opts.Count).
		Parse(args)
	if err != nil {
		return opts, nil, err
	}
	return opts, remainArgs, nil
}
