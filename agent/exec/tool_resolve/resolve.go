// Package toolresolve provides centralized binary resolution that respects
// both the system PATH, the well-known extra install paths (e.g. ~/.local/bin,
// ~/.opencode/bin), and user-configured extra paths from the terminal config.
//
// This package never modifies the process's PATH environment variable.
// Instead, LookPath dynamically searches the system PATH plus all extra paths.
// Callers spawning subprocesses should use AppendExtraPaths to build the env
// for the child process.
package tool_resolve

import (
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
)

// ExtraPaths are common install directories that may not be in the
// server process's PATH but where tools are commonly installed.
var ExtraPaths = []string{
	"/usr/local/bin",
	"/usr/local/go/bin",
}

func init() {
	if home, err := os.UserHomeDir(); err == nil {
		ExtraPaths = append(ExtraPaths,
			home+"/.local/bin",
			home+"/.opencode/bin",
			home+"/go/bin",
			home+"/.bun/bin",
			home+"/.fzf/bin",
		)
	}
	if out, err := exec.Command("npm", "bin", "-g").Output(); err == nil {
		npmBin := strings.TrimSpace(string(out))
		if npmBin != "" {
			ExtraPaths = append(ExtraPaths, npmBin)
		}
	}
	if bestNodeDir := findAllNodeVersionDirs(); len(bestNodeDir) > 0 {
		ExtraPaths = append(ExtraPaths, bestNodeDir...)
	}

	addNpmFromNodeDirs()
}

func addNpmFromNodeDirs() {
	for _, dir := range ExtraPaths {
		if strings.HasSuffix(dir, "/bin") {
			npmPath := filepath.Join(dir, "npm")
			if isExecutable(npmPath) {
				continue
			}
			bin2Dir := strings.TrimSuffix(dir, "/bin") + "/bin2"
			if info, err := os.Stat(bin2Dir); err == nil && info.IsDir() {
				npmPath := filepath.Join(bin2Dir, "npm")
				if isExecutable(npmPath) {
					ExtraPaths = append(ExtraPaths, bin2Dir)
					break
				}
			}
		}
	}

	knownNodeDirs := []string{
		"/consumerloan-codelensadmin/node/bin",
		"/consumerloan-codelensadmin/node/bin2",
		"/root/node/bin",
	}

	for _, dir := range knownNodeDirs {
		checkAndAddNpm(dir)
	}
}

func checkAndAddNpm(dir string) {
	if info, err := os.Stat(dir); err == nil && info.IsDir() {
		npmPath := filepath.Join(dir, "npm")
		if isExecutable(npmPath) {
			for _, p := range ExtraPaths {
				if p == dir {
					return
				}
			}
			ExtraPaths = append(ExtraPaths, dir)
		}
	}
}

type nodeVersionInfo struct {
	version string
	dir     string
}

func findAllNodeVersionDirs() []string {
	out, err := exec.Command("which", "-a", "node").Output()
	if err != nil {
		out2, err2 := exec.Command("which", "node").Output()
		if err2 != nil {
			return nil
		}
		out = out2
	}

	paths := strings.Split(strings.TrimSpace(string(out)), "\n")

	dirVersions := make(map[string]string)

	for _, path := range paths {
		path = strings.TrimSpace(path)
		if path == "" {
			continue
		}

		dir := filepath.Dir(path)

		versionOut, err := exec.Command(path, "--version").Output()
		if err != nil {
			continue
		}

		version := strings.TrimSpace(string(versionOut))
		version = strings.TrimPrefix(version, "v")

		if existingVersion, ok := dirVersions[dir]; !ok || CompareVersions(version, existingVersion) {
			dirVersions[dir] = version
		}
	}

	if len(dirVersions) == 0 {
		return nil
	}

	type dirVersion struct {
		dir     string
		version string
	}

	var sorted []dirVersion
	for dir, version := range dirVersions {
		sorted = append(sorted, dirVersion{dir: dir, version: version})
	}

	for i := 0; i < len(sorted)-1; i++ {
		for j := i + 1; j < len(sorted); j++ {
			if CompareVersions(sorted[j].version, sorted[i].version) {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}

	var result []string
	for _, dv := range sorted {
		result = append(result, dv.dir)
	}

	return result
}

func CompareVersions(v1, v2 string) bool {
	major1 := ExtractMajorVersion(v1)
	major2 := ExtractMajorVersion(v2)

	return major1 > major2
}

func ExtractMajorVersion(version string) int {
	version = strings.TrimPrefix(version, "v")
	parts := strings.Split(version, ".")
	if len(parts) == 0 {
		return 0
	}
	major, _ := strconv.Atoi(parts[0])
	return major
}

var (
	userPathsMu    sync.RWMutex
	userExtraPaths []string
)

func SetUserExtraPaths(paths []string) {
	userPathsMu.Lock()
	defer userPathsMu.Unlock()
	userExtraPaths = make([]string, len(paths))
	copy(userExtraPaths, paths)
}

func getUserExtraPaths() []string {
	userPathsMu.RLock()
	defer userPathsMu.RUnlock()
	result := make([]string, len(userExtraPaths))
	copy(result, userExtraPaths)
	return result
}

func AllExtraPaths() []string {
	result := make([]string, len(ExtraPaths))
	copy(result, ExtraPaths)
	result = append(result, getUserExtraPaths()...)
	return result
}

func GetFullSearchPATH() string {
	systemPath := os.Getenv("PATH")

	extras := AllExtraPaths()

	seen := make(map[string]bool)
	var orderedPaths []string

	for _, p := range strings.Split(systemPath, ":") {
		p = strings.TrimSpace(p)
		if p != "" && !seen[p] {
			seen[p] = true
			orderedPaths = append(orderedPaths, p)
		}
	}

	for _, p := range extras {
		p = strings.TrimSpace(p)
		if p != "" && !seen[p] {
			seen[p] = true
			orderedPaths = append(orderedPaths, p)
		}
	}

	type dirInfo struct {
		index       int
		dir         string
		nodeVersion string
		hasNode     bool
	}

	var dirInfos []dirInfo

	for idx, p := range orderedPaths {
		nodePath := filepath.Join(p, "node")
		info := dirInfo{index: idx, dir: p}

		if isExecutable(nodePath) {
			info.hasNode = true
			versionOut, err := exec.Command(nodePath, "--version").Output()
			if err == nil {
				version := strings.TrimSpace(string(versionOut))
				version = strings.TrimPrefix(version, "v")
				info.nodeVersion = version
			}
		}

		dirInfos = append(dirInfos, info)
	}

	sort.SliceStable(dirInfos, func(i, j int) bool {
		left := dirInfos[i]
		right := dirInfos[j]
		if left.hasNode != right.hasNode {
			return left.hasNode
		}
		if left.hasNode && right.hasNode && CompareVersions(left.nodeVersion, right.nodeVersion) != CompareVersions(right.nodeVersion, left.nodeVersion) {
			return CompareVersions(left.nodeVersion, right.nodeVersion)
		}
		return left.index < right.index
	})

	var result []string
	for _, info := range dirInfos {
		result = append(result, info.dir)
	}

	return strings.Join(result, ":")
}

func LookPath(name string) (string, error) {
	if strings.Contains(name, "/") {
		return lookPathDirect(name)
	}

	dirs := strings.Split(GetFullSearchPATH(), ":")
	for _, dir := range dirs {
		if dir == "" {
			continue
		}
		candidate := filepath.Join(dir, name)
		if isExecutable(candidate) {
			return candidate, nil
		}
	}
	return "", &lookPathError{name: name}
}

func IsAvailable(name string) bool {
	_, err := LookPath(name)
	return err == nil
}

func AppendExtraPaths(env []string) []string {
	fullPath := GetFullSearchPATH()

	for i, e := range env {
		if len(e) > 5 && e[:5] == "PATH=" {
			env[i] = "PATH=" + fullPath
			return env
		}
	}

	return append(env, "PATH="+fullPath)
}

func pathContains(pathVal, dir string) bool {
	for _, p := range strings.Split(pathVal, ":") {
		if p == dir {
			return true
		}
	}
	return false
}

func isExecutable(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	if info.IsDir() {
		return false
	}
	return info.Mode()&0111 != 0
}

func lookPathDirect(name string) (string, error) {
	if isExecutable(name) {
		return name, nil
	}
	return "", &lookPathError{name: name}
}

type lookPathError struct {
	name string
}

func (e *lookPathError) Error() string {
	return "executable file not found in PATH: " + e.name
}
