package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/xhd2015/less-flags"
	"github.com/xhd2015/agent-pro/agents/test-case-tree-runner/build"
	"github.com/xhd2015/agent-pro/agents/test-case-tree-runner/core"
	"github.com/xhd2015/agent-pro/agents/test-case-tree-runner/generate"
)

const usage = `Usage: test-case-tree-runner <command> [options]

Commands:
  test [-v] [--rm] [-count=N] <dir>
                             Run executable Go snippets from a test-case-tree directory.
  build [-v] [--rm] [--gen-dir DIR] [-count=N] <dir>
                             Validate generated snippets compile without executing behavior.
  generate-code [--dry-run] <dir>
                             Fill in missing Go code blocks using the test-case-tree-design-expert agent.
`

func main() {
	if err := runCli(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func runCli(args []string) error {
	if len(args) == 0 || args[0] == "-h" || args[0] == "--help" {
		fmt.Print(usage)
		return nil
	}
	switch args[0] {
	case "test":
		opts, cmdArgs, err := parseTestOptions(args[1:])
		if err != nil {
			return err
		}
		if len(cmdArgs) != 1 {
			return fmt.Errorf("test requires <dir>")
		}
		return TestTree(cmdArgs[0], opts)
	case "build":
		opts, cmdArgs, err := parseBuildOptions(args[1:])
		if err != nil {
			return err
		}
		if len(cmdArgs) != 1 {
			return fmt.Errorf("build requires <dir>")
		}
		return build.Build(cmdArgs[0], opts)
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

func parseTestOptions(args []string) (core.Options, []string, error) {
	opts := core.Options{Stderr: os.Stderr}
	remainArgs, err := lessflags.Bool("-v,--verbose", &opts.Verbose).
		Bool("--rm", &opts.RemoveTemp).
		Int("-count", &opts.Count).
		Parse(args)
	if err != nil {
		return opts, nil, err
	}
	return opts, remainArgs, nil
}

func parseBuildOptions(args []string) (core.Options, []string, error) {
	opts := core.Options{Stderr: os.Stderr}
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
