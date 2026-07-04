package codex

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/xhd2015/agent-pro/agent/debuglog"
)

const (
	npmPrefixTimeout   = 3 * time.Second
	loginShellTimeout  = 5 * time.Second
)

// codexInstallCandidates returns well-known codex binary locations for GUI-launched
// processes that inherit a minimal PATH (e.g. macOS Finder / menu-bar apps).
func codexInstallCandidates(home string) []string {
	if home == "" {
		return []string{
			"/opt/homebrew/bin/codex",
			"/usr/local/bin/codex",
		}
	}
	out := make([]string, 0, 12)
	if prefix, ok := parseNpmConfigPrefix(home); ok {
		out = append(out, filepath.Join(prefix, "bin", "codex"))
	}
	if version, ok := resolveNvmVersion(home); ok {
		out = append(out, filepath.Join(home, ".nvm", "versions", "node", version, "bin", "codex"))
	}
	out = append(out,
		filepath.Join(home, ".volta", "bin", "codex"),
		filepath.Join(home, ".local", "share", "fnm", "current", "bin", "codex"),
		filepath.Join(home, ".asdf", "shims", "codex"),
		filepath.Join(home, ".npm-global", "bin", "codex"),
		filepath.Join(home, "go", "bin", "codex"),
		"/opt/homebrew/bin/codex",
		"/usr/local/bin/codex",
	)
	return out
}

func npmExecutableCandidates(home string) []string {
	if home == "" {
		return []string{
			"/opt/homebrew/bin/npm",
			"/usr/local/bin/npm",
		}
	}
	out := []string{
		filepath.Join(home, ".npm-global", "bin", "npm"),
	}
	if version, ok := resolveNvmVersion(home); ok {
		out = append(out, filepath.Join(home, ".nvm", "versions", "node", version, "bin", "npm"))
	}
	out = append(out, "/opt/homebrew/bin/npm", "/usr/local/bin/npm")
	return out
}

func nodeExecutableCandidates(home, npmDir string) []string {
	out := make([]string, 0, 8)
	if home != "" {
		if version, ok := resolveNvmVersion(home); ok {
			out = append(out, filepath.Join(home, ".nvm", "versions", "node", version, "bin", "node"))
		}
	}
	if npmDir != "" && !isSystemToolDir(npmDir) {
		out = append(out, filepath.Join(npmDir, "node"))
	}
	out = append(out, "/opt/homebrew/bin/node", "/usr/local/bin/node")
	if npmDir != "" && isSystemToolDir(npmDir) {
		out = append(out, filepath.Join(npmDir, "node"))
	}
	return out
}

func isSystemToolDir(dir string) bool {
	dir = filepath.Clean(dir)
	return dir == "/opt/homebrew/bin" || dir == "/usr/local/bin"
}

func findFirstExecutable(candidates []string) (string, bool) {
	for _, candidate := range candidates {
		if isInvocableCLI(candidate) {
			return candidate, true
		}
	}
	return "", false
}

func findExecutableCodex(candidates []string) (string, bool) {
	return findFirstExecutable(candidates)
}

func probeCodexInstallPath() (string, bool) {
	home, homeErr := os.UserHomeDir()
	logCodexPathProbe("probe_start", map[string]any{
		"home":     home,
		"home_err": errString(homeErr),
		"path_env": os.Getenv("PATH"),
	})

	if homeErr != nil {
		if path, ok := probeFixedCodexPaths(""); ok {
			return path, true
		}
		return probeNpmPrefixCodex("")
	}

	if path, ok := probeFixedCodexPaths(home); ok {
		return path, true
	}
	if path, ok := probeNpmPrefixCodex(home); ok {
		return path, true
	}
	return probeCodexViaLoginShell(home)
}

func probeFixedCodexPaths(home string) (string, bool) {
	candidates := codexInstallCandidates(home)
	misses := make([]string, 0, len(candidates))
	for _, candidate := range candidates {
		if isInvocableCLI(candidate) {
			logCodexPathProbe("fixed_hit", map[string]any{
				"candidate": candidate,
			})
			return candidate, true
		}
		misses = append(misses, candidate)
	}
	logCodexPathProbe("fixed_miss", map[string]any{
		"candidates": misses,
	})
	return "", false
}

// resolveNvmVersion maps nvm default alias (e.g. "20") to an installed version dir (e.g. "v24.10.0").
func resolveNvmVersion(home string) (string, bool) {
	aliasPath := filepath.Join(home, ".nvm", "alias", "default")
	data, err := os.ReadFile(aliasPath)
	if err != nil {
		return "", false
	}
	alias := strings.TrimSpace(string(data))
	if alias == "" {
		return "", false
	}
	if strings.HasPrefix(alias, "v") || strings.HasPrefix(alias, "system") {
		return alias, true
	}
	nodeAliasPath := filepath.Join(home, ".nvm", "alias", "node", alias)
	if resolved, err := os.ReadFile(nodeAliasPath); err == nil {
		version := strings.TrimSpace(string(resolved))
		if version != "" {
			return version, true
		}
	}
	return alias, true
}

func parseNpmConfigPrefix(home string) (string, bool) {
	paths := []string{
		filepath.Join(home, ".npmrc"),
		filepath.Join(home, ".config", "npm", "npmrc"),
	}
	for _, path := range paths {
		if prefix, ok := readNpmRCPrefix(path); ok {
			logCodexPathProbe("npmrc_prefix", map[string]any{
				"file":   path,
				"prefix": prefix,
			})
			return prefix, true
		}
	}
	return "", false
}

func readNpmRCPrefix(path string) (string, bool) {
	f, err := os.Open(path)
	if err != nil {
		return "", false
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";") {
			continue
		}
		if strings.HasPrefix(line, "//") {
			continue
		}
		key, val, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		if strings.TrimSpace(key) != "prefix" {
			continue
		}
		prefix := strings.TrimSpace(val)
		prefix = strings.Trim(prefix, `"'`)
		if prefix == "" {
			return "", false
		}
		if strings.HasPrefix(prefix, "~/") {
			home, err := os.UserHomeDir()
			if err != nil {
				return "", false
			}
			prefix = filepath.Join(home, prefix[2:])
		}
		return filepath.Clean(prefix), true
	}
	return "", false
}

func probeNpmPrefixCodex(home string) (string, bool) {
	if prefix, ok := parseNpmConfigPrefix(home); ok {
		candidate := filepath.Join(prefix, "bin", "codex")
		if isInvocableCLI(candidate) {
			logCodexPathProbe("npmrc_hit", map[string]any{
				"prefix":    prefix,
				"candidate": candidate,
			})
			return candidate, true
		}
	}

	npmBin, npmOK := findNpmExecutable(home)
	if !npmOK {
		logCodexPathProbe("npm_prefix_miss", map[string]any{
			"reason": "npm_not_found",
		})
		return "", false
	}

	prefix, npmErr := runNpmPrefixGlobal(npmBin, home)
	if npmErr != nil {
		logCodexPathProbe("npm_prefix_miss", map[string]any{
			"npm_bin": npmBin,
			"reason":  "npm_prefix_failed",
			"error":   npmErr.Error(),
		})
		return "", false
	}
	logCodexPathProbe("npm_prefix_value", map[string]any{
		"npm_bin": npmBin,
		"prefix":  prefix,
	})
	if prefix == "" {
		logCodexPathProbe("npm_prefix_miss", map[string]any{
			"npm_bin": npmBin,
			"reason":  "empty_prefix",
		})
		return "", false
	}
	candidate := filepath.Join(prefix, "bin", "codex")
	if isInvocableCLI(candidate) {
		logCodexPathProbe("npm_prefix_hit", map[string]any{
			"npm_bin":   npmBin,
			"prefix":    prefix,
			"candidate": candidate,
		})
		return candidate, true
	}
	logCodexPathProbe("npm_prefix_miss", map[string]any{
		"npm_bin":   npmBin,
		"prefix":    prefix,
		"candidate": candidate,
		"reason":    "codex_not_invocable",
	})
	return "", false
}

func probeCodexViaLoginShell(home string) (string, bool) {
	user := os.Getenv("USER")
	if user == "" {
		user = os.Getenv("LOGNAME")
	}
	for _, shell := range []string{"bash", "zsh"} {
		path, ok := loginShellCommandV(shell, home, user, "command -v codex")
		if ok {
			logCodexPathProbe("login_shell_hit", map[string]any{
				"shell": shell,
				"path":  path,
			})
			return path, true
		}
	}
	logCodexPathProbe("login_shell_miss", map[string]any{
		"shells": []string{"bash", "zsh"},
	})
	return "", false
}

func loginShellCommandV(shell, home, user, command string) (string, bool) {
	ctx, cancel := context.WithTimeout(context.Background(), loginShellTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, shell, "-lic", command)
	cmd.Env = []string{
		"HOME=" + home,
		"USER=" + user,
		"LOGNAME=" + user,
		"PATH=/usr/bin:/bin:/usr/sbin:/sbin",
	}
	out, err := cmd.Output()
	if err != nil {
		logCodexPathProbe("login_shell_try", map[string]any{
			"shell": shell,
			"error": errString(err),
		})
		return "", false
	}
	path := strings.TrimSpace(string(out))
	if path == "" || !isInvocableCLI(path) {
		return "", false
	}
	return path, true
}

func runNpmPrefixGlobal(npmBin, home string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), npmPrefixTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, npmBin, "prefix", "-g")
	cmd.Env = minimalUserEnv(home)
	out, err := cmd.Output()
	if err == nil {
		logCodexPathProbe("npm_prefix_cmd", map[string]any{
			"npm_bin": npmBin,
			"via":     "direct",
		})
		return strings.TrimSpace(string(out)), nil
	}

	logCodexPathProbe("npm_prefix_cmd", map[string]any{
		"npm_bin":      npmBin,
		"via":          "node_fallback",
		"direct_error": err.Error(),
	})

	npmDir := filepath.Dir(npmBin)
	nodeBin, nodeOK := findNodeExecutable(home, npmDir)
	if !nodeOK {
		return "", fmt.Errorf("node not found for npm: %w", err)
	}
	npmCLI, cliErr := resolveNpmCLIEntrypoint(npmBin)
	if cliErr != nil {
		return "", fmt.Errorf("resolve npm cli: %w", cliErr)
	}

	cmd = exec.CommandContext(ctx, nodeBin, npmCLI, "prefix", "-g")
	cmd.Env = minimalUserEnv(home)
	out, err = cmd.Output()
	if err != nil {
		return "", fmt.Errorf("npm prefix via node: %w", err)
	}
	logCodexPathProbe("npm_prefix_cmd", map[string]any{
		"npm_bin":  npmBin,
		"node_bin": nodeBin,
		"npm_cli":  npmCLI,
		"via":      "node",
	})
	return strings.TrimSpace(string(out)), nil
}

func minimalUserEnv(home string) []string {
	user := os.Getenv("USER")
	if user == "" {
		user = os.Getenv("LOGNAME")
	}
	return []string{
		"HOME=" + home,
		"USER=" + user,
		"LOGNAME=" + user,
		"PATH=/usr/bin:/bin:/usr/sbin:/sbin",
	}
}

func resolveNpmCLIEntrypoint(npmBin string) (string, error) {
	target, err := os.Readlink(npmBin)
	if err != nil {
		return npmBin, nil
	}
	if filepath.IsAbs(target) {
		return target, nil
	}
	return filepath.Join(filepath.Dir(npmBin), target), nil
}

func findNpmExecutable(home string) (string, bool) {
	if path, ok := findFirstExecutable(npmExecutableCandidates(home)); ok {
		logCodexPathProbe("npm_bin", map[string]any{
			"npm_bin": path,
			"via":     "candidate",
		})
		return path, true
	}
	if path, err := exec.LookPath("npm"); err == nil && isInvocableCLI(path) {
		logCodexPathProbe("npm_bin", map[string]any{
			"npm_bin": path,
			"via":     "path_lookup",
		})
		return path, true
	}
	return "", false
}

func findNodeExecutable(home, npmDir string) (string, bool) {
	if path, ok := findFirstExecutable(nodeExecutableCandidates(home, npmDir)); ok {
		logCodexPathProbe("node_bin", map[string]any{
			"node_bin": path,
		})
		return path, true
	}
	if path, err := exec.LookPath("node"); err == nil && isInvocableCLI(path) {
		logCodexPathProbe("node_bin", map[string]any{
			"node_bin": path,
			"via":      "path_lookup",
		})
		return path, true
	}
	return "", false
}

// isInvocableCLI reports whether path is a runnable CLI shim (binary, symlink, or script).
func isInvocableCLI(path string) bool {
	info, err := os.Lstat(path)
	if err != nil || info.IsDir() {
		return false
	}
	if info.Mode()&0o111 != 0 {
		return true
	}
	if info.Mode()&os.ModeSymlink != 0 {
		if target, err := os.Stat(path); err == nil && !target.IsDir() {
			return true
		}
	}
	if runtime.GOOS == "windows" {
		return true
	}
	return false
}

func logCodexPathProbe(step string, fields map[string]any) {
	if fields == nil {
		fields = map[string]any{}
	}
	fields["step"] = step
	debuglog.Log(debuglog.Entry{
		Event: "codex_path_probe",
		Labels: map[string]string{
			"component": "codex_cli",
			"phase":     "resolve",
		},
		Fields: fields,
	})
}

func errString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}