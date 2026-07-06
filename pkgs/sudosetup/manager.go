package sudosetup

import (
	"encoding/json"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strings"
)

const (
	manifestFileName = "sudo-setup-manifest.json"
	sudoersDir       = "/etc/sudoers.d"
)

var (
	defaultFS     FS = osFS{}
	defaultRunner    = newExecRunner()
)

// Manager coordinates NOPASSWD sudoers setup for one command rule.
type Manager struct {
	Config Config
	Rule   Rule
	Runner Runner
	FS     FS
	// StdinIsTerminal overrides stdin TTY detection (tests only).
	StdinIsTerminal func() bool
}

func (m *Manager) fs() FS {
	if m.FS != nil {
		return m.FS
	}
	guardProductionDeps(m, "FS")
	return defaultFS
}

func (m *Manager) runner() Runner {
	if m.Runner != nil {
		return m.Runner
	}
	guardProductionDeps(m, "Runner")
	return defaultRunner
}

// SudoersPath returns the persistent sudoers drop-in path.
func (m *Manager) SudoersPath() string {
	return filepath.Join(sudoersDir, m.Config.SudoersName)
}

// ManifestPath returns the local install manifest path.
func (m *Manager) ManifestPath() string {
	cacheDir, err := m.fs().UserCacheDir()
	if err != nil {
		return filepath.Join("", m.Config.CacheDirName, manifestFileName)
	}
	return filepath.Join(cacheDir, m.Config.CacheDirName, manifestFileName)
}

// RenderSudoersLine returns the NOPASSWD line without a trailing newline.
func (m *Manager) RenderSudoersLine() (string, error) {
	username, err := m.resolveUsername()
	if err != nil {
		return "", err
	}
	line := username + " ALL=(root) NOPASSWD: " + m.Rule.Command
	if m.Rule.ArgsPattern != "" {
		line += " " + m.Rule.ArgsPattern
	}
	return line, nil
}

// IsInstalled checks persistent state via drop-in and manifest match.
func (m *Manager) IsInstalled() (bool, string) {
	return m.isInstalled()
}

// Detect reports install state plus live sudo cache and command probes.
func (m *Manager) Detect() Status {
	installed, installDetail := m.isInstalled()
	cacheWarm := m.isCacheWarm()
	canRun, canRunDetail := m.tryRunCommandNonInteractive()

	status := Status{
		Installed:              installed,
		InstallDetail:          installDetail,
		CacheWarm:              cacheWarm,
		CanRunNonInteractive:   canRun,
		CanRunDetail:           canRunDetail,
		Verdict:                m.computeVerdict(installed, cacheWarm, canRun),
	}
	return status
}

// EnsureInstalled installs the sudoers drop-in and manifest when not already present.
func (m *Manager) EnsureInstalled() error {
	if installed, _ := m.isInstalled(); installed {
		return nil
	}

	if !m.stdinIsTerminal() {
		return errInteractiveTerminalRequired("one-time sudo NOPASSWD setup")
	}

	username, err := m.resolveUsername()
	if err != nil {
		return err
	}

	line, err := m.RenderSudoersLine()
	if err != nil {
		return err
	}
	line += "\n"

	tmp, err := m.fs().CreateTemp("", "sudoers-*.tmp")
	if err != nil {
		return fmt.Errorf("temp sudoers file: %w", err)
	}
	tmpPath := tmp.Name()
	defer m.fs().Remove(tmpPath)

	if _, err := tmp.Write([]byte(line)); err != nil {
		tmp.Close()
		return fmt.Errorf("write temp sudoers: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return err
	}

	if err := m.runner().Run("sudo", "visudo", "-cf", tmpPath); err != nil {
		return fmt.Errorf("visudo validation failed: %w", err)
	}

	sudoersPath := m.SudoersPath()
	if err := m.runner().Run("sudo", "install", "-o", "root", "-g", "wheel", "-m", "0440", tmpPath, sudoersPath); err != nil {
		return fmt.Errorf("install sudoers drop-in: %w", err)
	}
	if err := m.fs().WriteFile(sudoersPath, []byte(line), 0o440); err != nil {
		return fmt.Errorf("write sudoers drop-in: %w", err)
	}

	if err := m.writeManifest(username); err != nil {
		return fmt.Errorf("write install manifest: %w", err)
	}
	return nil
}

// Remove deletes the sudoers drop-in and manifest and flushes the sudo cache.
func (m *Manager) Remove() error {
	fs := m.fs()
	sudoersPath := m.SudoersPath()

	if _, err := fs.Stat(sudoersPath); err == nil {
		if !m.stdinIsTerminal() {
			return errInteractiveTerminalRequired("sudo NOPASSWD removal")
		}
		if err := m.runner().Run("sudo", "rm", "-f", sudoersPath); err != nil {
			return fmt.Errorf("sudo rm %s: %w", sudoersPath, err)
		}
		_ = fs.Remove(sudoersPath)
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("stat %s: %w", sudoersPath, err)
	}

	if err := m.removeManifest(); err != nil {
		return err
	}

	if err := m.runner().Run("sudo", "-k"); err != nil {
		return fmt.Errorf("sudo -k: %w", err)
	}
	return nil
}

func (m *Manager) isInstalled() (bool, string) {
	fs := m.fs()
	sudoersPath := m.SudoersPath()

	if _, err := fs.Stat(sudoersPath); err != nil {
		if os.IsNotExist(err) {
			return false, "sudoers drop-in not present"
		}
		return false, fmt.Sprintf("stat sudoers drop-in: %v", err)
	}

	manifest, err := m.readManifest()
	if err != nil {
		if os.IsNotExist(err) {
			return false, "install manifest missing (drop-in may be orphaned)"
		}
		return false, fmt.Sprintf("read manifest: %v", err)
	}

	username, err := m.resolveUsername()
	if err != nil {
		return false, err.Error()
	}
	if manifest.Username != username {
		return false, fmt.Sprintf("manifest user %q != current user %q", manifest.Username, username)
	}
	if manifest.Command != m.Rule.Command {
		return false, fmt.Sprintf("manifest command %q != current %q", manifest.Command, m.Rule.Command)
	}
	if manifest.ArgsPattern != m.Rule.ArgsPattern {
		return false, fmt.Sprintf("manifest args_pattern %q != current %q", manifest.ArgsPattern, m.Rule.ArgsPattern)
	}
	return true, fmt.Sprintf("drop-in %s matches manifest for %s", sudoersPath, username)
}

func (m *Manager) isCacheWarm() bool {
	return m.runner().Run("sudo", "-n", "true") == nil
}

func (m *Manager) tryRunCommandNonInteractive() (bool, string) {
	args := m.probeCommandArgs()
	_, err := m.runner().CombinedOutput("sudo", args...)
	if err == nil {
		return true, "command succeeded non-interactively"
	}
	msg := strings.TrimSpace(err.Error())
	if msg == "" {
		msg = "sudo: a password is required"
	}
	return false, msg
}

func (m *Manager) probeCommandArgs() []string {
	args := []string{"-n", m.Rule.Command}
	if m.Rule.ArgsPattern == "" {
		return args
	}
	for _, part := range strings.Fields(m.Rule.ArgsPattern) {
		part = strings.ReplaceAll(part, "*", "/dev/null")
		args = append(args, part)
	}
	return args
}

func (m *Manager) computeVerdict(installed, cacheWarm, canRun bool) string {
	switch {
	case installed && canRun:
		return "permanent NOPASSWD is configured"
	case !installed && canRun && cacheWarm:
		return "command works only via sudo cache (rule not installed)"
	case !installed && canRun:
		return "command works without rule (likely sudo cache)"
	case !installed && cacheWarm:
		return "sudo timestamp cache warm (rule not installed)"
	default:
		return "password required"
	}
}

func (m *Manager) resolveUsername() (string, error) {
	if name := strings.TrimSpace(m.Config.Username); name != "" {
		return name, nil
	}
	if name := strings.TrimSpace(os.Getenv("SUDO_USER")); name != "" && name != "root" {
		return name, nil
	}
	if name := strings.TrimSpace(os.Getenv("USER")); name != "" {
		return name, nil
	}
	u, err := user.Current()
	if err != nil {
		return "", fmt.Errorf("current user: %w", err)
	}
	return u.Username, nil
}

func (m *Manager) readManifest() (*Manifest, error) {
	data, err := m.fs().ReadFile(m.ManifestPath())
	if err != nil {
		return nil, err
	}
	var manifest Manifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return nil, fmt.Errorf("parse manifest: %w", err)
	}
	return &manifest, nil
}

func (m *Manager) writeManifest(username string) error {
	cacheDir, err := m.fs().UserCacheDir()
	if err != nil {
		return err
	}
	dir := filepath.Join(cacheDir, m.Config.CacheDirName)
	if err := m.fs().MkdirAll(dir, 0o700); err != nil {
		return err
	}

	data, err := json.MarshalIndent(Manifest{
		Username:    username,
		Command:     m.Rule.Command,
		ArgsPattern: m.Rule.ArgsPattern,
	}, "", "  ")
	if err != nil {
		return err
	}
	return m.fs().WriteFile(m.ManifestPath(), data, 0o600)
}

func (m *Manager) removeManifest() error {
	err := m.fs().Remove(m.ManifestPath())
	if os.IsNotExist(err) {
		return nil
	}
	return err
}