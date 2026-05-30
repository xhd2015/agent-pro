package tool_exec

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/xhd2015/agent-pro/agent/exec/tool_resolve"
)

type Command struct {
	*exec.Cmd
}

type Options struct {
	CustomPath string
	Env        map[string]string
	Dir        string
}

func New(binary string, args []string, opts *Options) (*Command, error) {
	if opts == nil {
		opts = &Options{}
	}

	binaryPath, err := resolveBinaryPath(binary, opts.CustomPath)
	if err != nil {
		return nil, err
	}

	cmd := exec.Command(binaryPath, args...)

	cmd.Env = setupEnvironment(os.Environ(), opts.Env)

	if opts.Dir != "" {
		cmd.Dir = opts.Dir
	}

	return &Command{Cmd: cmd}, nil
}

func resolveBinaryPath(binary, customPath string) (string, error) {
	if customPath != "" {
		if !filepath.IsAbs(customPath) {
			absPath, err := filepath.Abs(customPath)
			if err == nil {
				customPath = absPath
			}
		}

		if info, err := os.Stat(customPath); err == nil && !info.IsDir() {
			if info.Mode()&0111 != 0 {
				return customPath, nil
			}
		}
	}

	if filepath.IsAbs(binary) {
		return binary, nil
	}

	return tool_resolve.LookPath(binary)
}

func setupEnvironment(baseEnv []string, extraVars map[string]string) []string {
	env := make([]string, len(baseEnv))
	copy(env, baseEnv)

	env = tool_resolve.AppendExtraPaths(env)

	if len(extraVars) > 0 {
		env = setEnvVars(env, extraVars)
	}

	return env
}

func setEnvVars(env []string, vars map[string]string) []string {
	envMap := make(map[string]string)
	for _, e := range env {
		for i := 0; i < len(e); i++ {
			if e[i] == '=' {
				key := e[:i]
				value := e[i+1:]
				envMap[key] = value
				break
			}
		}
	}

	for k, v := range vars {
		envMap[k] = v
	}

	result := make([]string, 0, len(envMap))
	for k, v := range envMap {
		result = append(result, fmt.Sprintf("%s=%s", k, v))
	}

	return result
}

func MustNew(binary string, args []string, opts *Options) *Command {
	cmd, err := New(binary, args, opts)
	if err != nil {
		panic(fmt.Sprintf("toolexec: failed to create command for %s: %v", binary, err))
	}
	return cmd
}

func IsAvailable(binary, customPath string) bool {
	_, err := resolveBinaryPath(binary, customPath)
	return err == nil
}
