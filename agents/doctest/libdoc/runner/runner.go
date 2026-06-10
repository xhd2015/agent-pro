package runner

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/xhd2015/less-flags"
	runnerbuild "github.com/xhd2015/agent-pro/agents/test-case-tree-runner/build"
	"github.com/xhd2015/agent-pro/agents/test-case-tree-runner/core"
)

var ErrNoTestsFound = errors.New("no tests found")

func Build(dir string) error {
	return runnerbuild.Build(dir, core.Options{RemoveTemp: false})
}

func BuildArgs(args []string) error {
	opts, remainArgs, err := parseBuildOptions(args)
	if err != nil {
		return err
	}
	if len(remainArgs) != 1 {
		return fmt.Errorf("build requires <dir>")
	}
	targetDir, _ := filepath.Abs(remainArgs[0])
	root, ok := ResolveRoot(targetDir)
	if !ok {
		return ErrNoTestsFound
	}
	opts.SubDir = targetDir
	err = runnerbuild.Build(root, opts)
	if err != nil && strings.Contains(err.Error(), "no runnable test cases found") {
		return ErrNoTestsFound
	}
	return err
}

func Test(args []string) error {
	opts, remainArgs, err := parseTestOptions(args)
	if err != nil {
		return err
	}
	if len(remainArgs) != 1 {
		return fmt.Errorf("test requires <dir>")
	}
	targetDir, _ := filepath.Abs(remainArgs[0])
	root, ok := ResolveRoot(targetDir)
	if !ok {
		return ErrNoTestsFound
	}
	opts.SubDir = targetDir
	err = runnerbuild.Test(root, opts)
	if err != nil && strings.Contains(err.Error(), "no runnable test cases found") {
		return ErrNoTestsFound
	}
	return err
}

func parseBuildOptions(args []string) (core.Options, []string, error) {
	opts := core.Options{Stderr: os.Stderr, RemoveTemp: false}
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
	opts := core.Options{Stderr: os.Stderr, RemoveTemp: false}
	remainArgs, err := lessflags.Bool("-v,--verbose", &opts.Verbose).
		Bool("--rm", &opts.RemoveTemp).
		Int("-count", &opts.Count).
		Parse(args)
	if err != nil {
		return core.Options{}, nil, err
	}
	return opts, remainArgs, nil
}
