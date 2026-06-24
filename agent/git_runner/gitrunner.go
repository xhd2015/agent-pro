package git_runner

import (
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/xhd2015/agent-pro/agent/exec/tool_resolve"
)

var isToolAvailable = tool_resolve.IsAvailable

type SSHKeyConfig struct {
	KeyPath  string
	ProxyURL string
}

type Command struct {
	args      []string
	dir       string
	sshConfig *SSHKeyConfig
	env       map[string]string
	noPrompts bool
}

func NewCommand(args ...string) *Command {
	return &Command{
		args:      args,
		env:       make(map[string]string),
		noPrompts: true,
	}
}

func (c *Command) Dir(dir string) *Command {
	c.dir = dir
	return c
}

func (c *Command) WithSSHKey(keyPath string) *Command {
	c.sshConfig = &SSHKeyConfig{KeyPath: keyPath}
	return c
}

func (c *Command) WithSSHConfig(cfg *SSHKeyConfig) *Command {
	c.sshConfig = cfg
	return c
}

func (c *Command) WithEnv(key, value string) *Command {
	c.env[key] = value
	return c
}

func (c *Command) AllowPrompts() *Command {
	c.noPrompts = false
	return c
}

func (c *Command) Validate() error {
	if c == nil {
		return nil
	}
	return validateSSHConfig(c.sshConfig)
}

func (c *Command) Build() *exec.Cmd {
	cmd := exec.Command("git", c.args...)
	if c.dir != "" {
		cmd.Dir = c.dir
	}

	env := os.Environ()

	customKeys := make(map[string]bool)
	for k := range c.env {
		customKeys[k] = true
	}

	if c.sshConfig != nil {
		sshCmd := buildSSHCommand(c.sshConfig)
		if sshCmd != "" {
			customKeys["GIT_SSH_COMMAND"] = true
		}
	}

	if c.noPrompts {
		if _, ok := c.env["GIT_ASKPASS"]; !ok {
			customKeys["GIT_ASKPASS"] = true
		}
		if _, ok := c.env["GIT_TERMINAL_PROMPT"]; !ok {
			customKeys["GIT_TERMINAL_PROMPT"] = true
		}
		if _, ok := c.env["SSH_ASKPASS_REQUIRE"]; !ok {
			customKeys["SSH_ASKPASS_REQUIRE"] = true
		}
	}

	env = overrideEnv(env, customKeys)

	for k, v := range c.env {
		env = append(env, fmt.Sprintf("%s=%s", k, v))
	}

	if c.sshConfig != nil {
		sshCmd := buildSSHCommand(c.sshConfig)
		if sshCmd != "" {
			env = append(env, fmt.Sprintf("GIT_SSH_COMMAND=%s", sshCmd))
		}
	}

	if c.noPrompts {
		if _, ok := c.env["GIT_ASKPASS"]; !ok {
			env = append(env, "GIT_ASKPASS=/bin/true")
		}
		if _, ok := c.env["GIT_TERMINAL_PROMPT"]; !ok {
			env = append(env, "GIT_TERMINAL_PROMPT=0")
		}
		if _, ok := c.env["SSH_ASKPASS_REQUIRE"]; !ok {
			env = append(env, "SSH_ASKPASS_REQUIRE=never")
		}
	}

	cmd.Env = env
	cmd.Stdin = nil
	return cmd
}

func overrideEnv(env []string, keys map[string]bool) []string {
	if len(keys) == 0 {
		return env
	}
	filtered := make([]string, 0, len(env))
	for _, e := range env {
		k := envKey(e)
		if k == "" || !keys[k] {
			filtered = append(filtered, e)
		}
	}
	return filtered
}

func envKey(entry string) string {
	for i := 0; i < len(entry); i++ {
		if entry[i] == '=' {
			return entry[:i]
		}
	}
	return ""
}

func buildSSHCommand(cfg *SSHKeyConfig) string {
	if cfg == nil {
		return ""
	}
	args := []string{
		"ssh",
		"-o", "StrictHostKeyChecking=no",
		"-o", "UserKnownHostsFile=/dev/null",
		"-o", "BatchMode=yes",
	}
	if cfg.KeyPath != "" {
		args = append(args, "-i", cfg.KeyPath)
	}
	if proxyCmd := proxyCommandForURL(cfg.ProxyURL); proxyCmd != "" {
		args = append(args, "-o", "ProxyCommand="+proxyCmd)
	}
	return joinShellArgs(args)
}

func proxyCommandForURL(rawURL string) string {
	rawURL = strings.TrimSpace(rawURL)
	if rawURL == "" {
		return ""
	}
	u, err := url.Parse(rawURL)
	if err != nil {
		return ""
	}
	host := u.Host
	if host == "" {
		return ""
	}

	proto := strings.ToLower(u.Scheme)
	if proto == "" {
		proto = "http"
	}

	proxyVersion := ""
	switch proto {
	case "http":
		proxyVersion = "connect"
	case "socks5":
		proxyVersion = "5"
	default:
		return ""
	}

	return joinShellArgs([]string{
		"nc",
		"-X", proxyVersion,
		"-x", host,
		"%h",
		"%p",
	})
}

func validateSSHConfig(cfg *SSHKeyConfig) error {
	if cfg == nil {
		return nil
	}
	if proxyCommandForURL(cfg.ProxyURL) == "" {
		return nil
	}
	if isToolAvailable("nc") {
		return nil
	}
	return fmt.Errorf("ssh proxy requires 'nc' on the remote server, but it is not installed. Install it first, for example: `remote-agent exec apt install netcat`")
}

func joinShellArgs(args []string) string {
	var quoted []string
	for _, arg := range args {
		if arg == "" {
			continue
		}
		quoted = append(quoted, shellQuote(arg))
	}
	return strings.Join(quoted, " ")
}

func shellQuote(s string) string {
	if s == "" {
		return "''"
	}
	return strconv.Quote(s)
}

func (c *Command) Run() ([]byte, error) {
	cmd := c.Build()
	return cmd.CombinedOutput()
}

func (c *Command) Output() ([]byte, error) {
	cmd := c.Build()
	return cmd.Output()
}

func (c *Command) RunSilent() error {
	cmd := c.Build()
	cmd.Stdout = nil
	cmd.Stderr = nil
	return cmd.Run()
}

func (c *Command) Exec() *exec.Cmd {
	return c.Build()
}

func EnsureAvailable() error {
	if !tool_resolve.IsAvailable("git") {
		return fmt.Errorf("git is not installed. Please install git first (e.g. apt-get update && apt-get install -y git)")
	}
	return nil
}

func Clone(repoURL, targetDir string, sshKeyPath ...string) *Command {
	cmd := NewCommand("clone", "--progress", repoURL, targetDir)
	if len(sshKeyPath) > 0 && sshKeyPath[0] != "" {
		cmd.WithSSHKey(sshKeyPath[0])
	}
	return cmd
}

func Fetch(sshKeyPath ...string) *Command {
	cmd := NewCommand("fetch", "--progress")
	if len(sshKeyPath) > 0 && sshKeyPath[0] != "" {
		cmd.WithSSHKey(sshKeyPath[0])
	}
	return cmd
}

func Pull(sshKeyPath ...string) *Command {
	cmd := NewCommand("pull", "--progress")
	if len(sshKeyPath) > 0 && sshKeyPath[0] != "" {
		cmd.WithSSHKey(sshKeyPath[0])
	}
	return cmd
}

func PullFFOnly(sshKeyPath ...string) *Command {
	cmd := NewCommand("pull", "--ff-only")
	if len(sshKeyPath) > 0 && sshKeyPath[0] != "" {
		cmd.WithSSHKey(sshKeyPath[0])
	}
	return cmd
}

func Push(branch string, sshKeyPath ...string) *Command {
	cmd := NewCommand("push", "origin", fmt.Sprintf("HEAD:%s", branch), "--progress")
	if len(sshKeyPath) > 0 && sshKeyPath[0] != "" {
		cmd.WithSSHKey(sshKeyPath[0])
	}
	return cmd
}

func Add(paths ...string) *Command {
	args := append([]string{"add"}, paths...)
	return NewCommand(args...)
}

func Reset(paths ...string) *Command {
	args := append([]string{"reset", "HEAD"}, paths...)
	return NewCommand(args...)
}

func Commit(message string) *Command {
	return NewCommand("commit", "-m", message)
}

func IndexLockPath(dir string) (string, error) {
	out, err := NewCommand("rev-parse", "--git-path", "index.lock").Dir(dir).Output()
	if err != nil {
		return "", fmt.Errorf("resolve index.lock path: %w", err)
	}
	lockPath := strings.TrimSpace(string(out))
	if lockPath == "" {
		return "", fmt.Errorf("resolve index.lock path: empty path")
	}
	if !filepath.IsAbs(lockPath) {
		lockPath = filepath.Join(dir, lockPath)
	}
	return filepath.Clean(lockPath), nil
}

func RemoveStaleIndexLock(dir string) error {
	lockPath, err := IndexLockPath(dir)
	if err != nil {
		return err
	}
	if err := os.Remove(lockPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("remove stale index.lock %s: %w", lockPath, err)
	}
	return nil
}

func CommitWithRetry(dir, message string, maxAttempts int) ([]byte, error) {
	if maxAttempts < 1 {
		maxAttempts = 1
	}
	var lastOutput []byte
	var lastErr error
	for attempt := 0; attempt < maxAttempts; attempt++ {
		if attempt > 0 {
			time.Sleep(time.Duration(attempt) * 50 * time.Millisecond)
		}
		_ = RemoveStaleIndexLock(dir)
		output, err := Commit(message).Dir(dir).Run()
		if err == nil {
			return output, nil
		}
		lastOutput = output
		lastErr = err
		if !isIndexLockContention(string(output), err) {
			return output, err
		}
	}
	return lastOutput, lastErr
}

func isIndexLockContention(output string, err error) bool {
	combined := output
	if err != nil {
		combined += "\n" + err.Error()
	}
	return strings.Contains(combined, "index.lock")
}

func Diff(args ...string) *Command {
	return NewCommand(append([]string{"diff"}, args...)...)
}

func DiffCached() *Command {
	return NewCommand("diff", "--cached")
}

func Status(args ...string) *Command {
	return NewCommand(append([]string{"status"}, args...)...)
}

func Branch(args ...string) *Command {
	return NewCommand(append([]string{"branch"}, args...)...)
}

func RevParse(args ...string) *Command {
	return NewCommand(append([]string{"rev-parse"}, args...)...)
}

func ForEachRef(args ...string) *Command {
	return NewCommand(append([]string{"for-each-ref"}, args...)...)
}

func LsFiles(args ...string) *Command {
	return NewCommand(append([]string{"ls-files"}, args...)...)
}

func Show(args ...string) *Command {
	return NewCommand(append([]string{"show"}, args...)...)
}

func Config(key, value string) *Command {
	return NewCommand("config", key, value)
}

func IsRepo(dir string) bool {
	cmd := NewCommand("rev-parse", "--git-dir").Dir(dir)
	return cmd.RunSilent() == nil
}

func GetCurrentBranch(dir string) (string, error) {
	cmd := NewCommand("branch", "--show-current").Dir(dir)
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

func Checkout(paths ...string) *Command {
	args := append([]string{"checkout", "--"}, paths...)
	return NewCommand(args...)
}

func CheckIgnore(path string) *Command {
	return NewCommand("check-ignore", "-q", path)
}

func IsIgnored(dir, path string) bool {
	cmd := CheckIgnore(path).Dir(dir)
	return cmd.RunSilent() == nil
}
