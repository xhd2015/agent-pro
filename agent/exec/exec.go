package exec

import (
	"context"
	"fmt"
	"os"
	osexec "os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

type PathsConfig struct {
	RootDirName string
	DataDirName string
	BinDirName  string
}

type Dirs struct {
	Root     string
	DataRoot string
}

func (c *PathsConfig) HomeRoot() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("get home directory: %w", err)
	}
	return filepath.Join(home, c.RootDirName), nil
}

func (c *PathsConfig) DefaultRoot() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("get working directory: %w", err)
	}
	return filepath.Join(cwd, c.RootDirName), nil
}

func (c *PathsConfig) HomeDataRoot() (string, error) {
	root, err := c.HomeRoot()
	if err != nil {
		return "", err
	}
	return c.DataRoot(root), nil
}

func (c *PathsConfig) DefaultDataRoot() (string, error) {
	root, err := c.DefaultRoot()
	if err != nil {
		return "", err
	}
	return c.DataRoot(root), nil
}

func (c *PathsConfig) Resolve(rootFlag string) (Dirs, error) {
	if strings.TrimSpace(rootFlag) != "" {
		dataRoot, err := c.ResolveDataRoot(rootFlag)
		if err != nil {
			return Dirs{}, err
		}
		root, err := c.RootFromDataRoot(dataRoot)
		if err != nil {
			return Dirs{}, err
		}
		return Dirs{Root: root, DataRoot: dataRoot}, nil
	}
	root, err := c.HomeRoot()
	if err != nil {
		return Dirs{}, err
	}
	return Dirs{Root: root, DataRoot: c.DataRoot(root)}, nil
}

func (c *PathsConfig) ResolveDataRoot(rootFlag string) (string, error) {
	if strings.TrimSpace(rootFlag) == "" {
		return c.HomeDataRoot()
	}
	dataRoot, err := filepath.Abs(filepath.Clean(rootFlag))
	if err != nil {
		return "", fmt.Errorf("resolve root: %w", err)
	}
	return dataRoot, nil
}

func (c *PathsConfig) RootFromDataRoot(dataRoot string) (string, error) {
	if strings.TrimSpace(dataRoot) == "" {
		return c.HomeRoot()
	}
	abs, err := filepath.Abs(filepath.Clean(dataRoot))
	if err != nil {
		return "", fmt.Errorf("resolve data root: %w", err)
	}
	return filepath.Dir(abs), nil
}

func (c *PathsConfig) DataRoot(root string) string {
	return filepath.Join(filepath.Clean(root), c.DataDirName)
}

func (c *PathsConfig) BinDir(root string) string {
	return filepath.Join(filepath.Clean(root), c.BinDirName)
}

func (c *PathsConfig) HomeBinDir() (string, error) {
	root, err := c.HomeRoot()
	if err != nil {
		return "", err
	}
	return c.BinDir(root), nil
}

func (c *PathsConfig) BinDirForDataRoot(dataRoot string) (string, error) {
	root, err := c.RootFromDataRoot(dataRoot)
	if err != nil {
		return "", err
	}
	return c.BinDir(root), nil
}

type Env struct {
	Paths            *PathsConfig
	ConfigRootEnvKey string
}

func NewEnv(paths *PathsConfig, configRootEnvKey string) *Env {
	return &Env{
		Paths:            paths,
		ConfigRootEnvKey: configRootEnvKey,
	}
}

func (e *Env) CommandContext(ctx context.Context, name string, args ...string) *osexec.Cmd {
	if resolved, err := e.LookPath(name); err == nil {
		name = resolved
	}
	cmd := osexec.CommandContext(ctx, name, args...)
	cmd.Env = e.Environ()
	return cmd
}

func (e *Env) Command(name string, args ...string) *osexec.Cmd {
	if resolved, err := e.LookPath(name); err == nil {
		name = resolved
	}
	cmd := osexec.Command(name, args...)
	cmd.Env = e.Environ()
	return cmd
}

func (e *Env) LookPath(file string) (string, error) {
	if strings.TrimSpace(file) == "" {
		return "", fmt.Errorf("empty executable name")
	}
	if hasPathSeparator(file) {
		return osexec.LookPath(file)
	}
	if path, ok := e.lookPathInPATH(file, envValue(e.Environ(), "PATH")); ok {
		return path, nil
	}
	return osexec.LookPath(file)
}

func (e *Env) Environ() []string {
	return e.WithConfigRoot(e.WithBinPath(os.Environ()))
}

func (e *Env) PrependBinToPATH() error {
	binDir := e.homeBinDir()
	if strings.TrimSpace(binDir) == "" {
		return nil
	}
	oldPath := os.Getenv("PATH")
	newPath := binDir
	if oldPath != "" {
		newPath += string(os.PathListSeparator) + oldPath
	}
	return os.Setenv("PATH", newPath)
}

func (e *Env) PrependBinToPATHForDataRoot(dataRoot string) error {
	binDir, err := e.Paths.BinDirForDataRoot(dataRoot)
	if err != nil {
		return err
	}
	if strings.TrimSpace(binDir) == "" {
		return nil
	}
	oldPath := os.Getenv("PATH")
	newPath := binDir
	if oldPath != "" {
		newPath += string(os.PathListSeparator) + oldPath
	}
	return os.Setenv("PATH", newPath)
}

func (e *Env) WithBinPath(env []string) []string {
	env = append([]string(nil), env...)
	binDir := e.homeBinDir()
	if strings.TrimSpace(binDir) == "" {
		return env
	}
	return e.upsertPath(env, binDir)
}

func (e *Env) WithConfigRoot(env []string) []string {
	env = append([]string(nil), env...)
	configDir := e.configRootDir()
	if strings.TrimSpace(configDir) == "" {
		return env
	}
	return upsertEnv(env, e.ConfigRootEnvKey, configDir)
}

func (e *Env) homeBinDir() string {
	binDir, err := e.Paths.HomeBinDir()
	if err != nil {
		return ""
	}
	return binDir
}

func (e *Env) configRootDir() string {
	root, err := e.Paths.HomeRoot()
	if err != nil {
		return ""
	}
	return filepath.Join(root, "config")
}

func (e *Env) upsertPath(env []string, dir string) []string {
	path := envValue(env, "PATH")
	values := []string{dir}
	for _, part := range filepath.SplitList(path) {
		if part == "" || samePath(part, dir) {
			continue
		}
		values = append(values, part)
	}
	return upsertEnv(env, "PATH", strings.Join(values, string(os.PathListSeparator)))
}

func (e *Env) lookPathInPATH(file string, path string) (string, bool) {
	for _, dir := range filepath.SplitList(path) {
		if dir == "" {
			dir = "."
		}
		candidate := filepath.Join(dir, file)
		if isExecutable(candidate) {
			return candidate, true
		}
		if runtime.GOOS == "windows" {
			for _, ext := range []string{".exe", ".cmd", ".bat", ".com"} {
				if isExecutable(candidate + ext) {
					return candidate + ext, true
				}
			}
		}
	}
	return "", false
}

func upsertEnv(env []string, key string, value string) []string {
	prefix := key + "="
	out := make([]string, 0, len(env)+1)
	last := -1
	for i, entry := range env {
		if strings.HasPrefix(entry, prefix) {
			last = i
		}
	}
	if last == -1 {
		return append(env, prefix+value)
	}
	for i, entry := range env {
		if strings.HasPrefix(entry, prefix) {
			if i == last {
				out = append(out, prefix+value)
			}
			continue
		}
		out = append(out, entry)
	}
	return out
}

func envValue(env []string, key string) string {
	prefix := key + "="
	for i := len(env) - 1; i >= 0; i-- {
		entry := env[i]
		if strings.HasPrefix(entry, prefix) {
			return strings.TrimPrefix(entry, prefix)
		}
	}
	return ""
}

func hasPathSeparator(file string) bool {
	return strings.ContainsRune(file, os.PathSeparator) || (os.PathSeparator != '/' && strings.Contains(file, "/"))
}

func isExecutable(path string) bool {
	info, err := os.Stat(path)
	if err != nil || info.IsDir() {
		return false
	}
	if runtime.GOOS == "windows" {
		return true
	}
	return info.Mode()&0o111 != 0
}

func samePath(a, b string) bool {
	return filepath.Clean(a) == filepath.Clean(b)
}