package runner

import (
	"fmt"
	"os"

	"github.com/xhd2015/less-flags"
	runnerbuild "github.com/xhd2015/agent-pro/agents/test-case-tree-runner/build"
	"github.com/xhd2015/agent-pro/agents/test-case-tree-runner/core"
)

func Build(dir string) error {
	return runnerbuild.Build(dir, core.Options{RemoveTemp: true})
}

func BuildArgs(args []string) error {
	opts, remainArgs, err := parseBuildOptions(args)
	if err != nil {
		return err
	}
	if len(remainArgs) != 1 {
		return fmt.Errorf("build requires <dir>")
	}
	return runnerbuild.Build(remainArgs[0], opts)
}

func Test(args []string) error {
	opts, remainArgs, err := parseTestOptions(args)
	if err != nil {
		return err
	}
	if len(remainArgs) != 1 {
		return fmt.Errorf("test requires <dir>")
	}
	return runnerbuild.Test(remainArgs[0], opts)
}

func parseBuildOptions(args []string) (core.Options, []string, error) {
	opts := core.Options{Stderr: os.Stderr, RemoveTemp: true}
	remainArgs, err := lessflags.Bool("-v,--verbose", &opts.Verbose).
		Bool("--rm", &opts.RemoveTemp).
		String("--gen-dir", &opts.GenDir).
		Int("-count", &opts.Count).
		Parse(args)
	if err != nil {
		return core.Options{}, nil, err
	}
	return opts, remainArgs, nil
}

func parseTestOptions(args []string) (core.Options, []string, error) {
	opts := core.Options{Stderr: os.Stderr, RemoveTemp: true}
	remainArgs, err := lessflags.Bool("-v,--verbose", &opts.Verbose).
		Bool("--rm", &opts.RemoveTemp).
		Int("-count", &opts.Count).
		Parse(args)
	if err != nil {
		return core.Options{}, nil, err
	}
	return opts, remainArgs, nil
}
