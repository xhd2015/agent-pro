package commit_msg

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/xhd2015/dot-pkgs/go-pkgs/file/detect"
	"github.com/xhd2015/dot-pkgs/go-pkgs/git/submodule"
	"github.com/xhd2015/gitops/git"
	"github.com/xhd2015/gitops/gitwrite"
)

// Kinds reported by DetectUnstageCandidates.
const (
	UnstageKindBinary    = "binary"
	UnstageKindSubmodule = "submodule"
)

// UnstageCandidate is one path that gen-commit-msg removes from the index before
// generating a message: a staged binary file, or any staged path inside a
// submodule. Callers that stage files on the user's behalf can ask
// DetectUnstageCandidates what generation would drop and keep those paths out of
// the index, so generation never has to mutate a freshly prepared staging area.
type UnstageCandidate struct {
	// Path is repo-relative.
	Path string
	// Kind is UnstageKindBinary or UnstageKindSubmodule.
	Kind string
	// Desc describes the file type; only set for binaries.
	Desc string
	// SubmoduleDir is the owning submodule; only set for submodule paths.
	SubmoduleDir string
}

// UnstageCandidatePaths returns the paths of candidates, preserving order.
func UnstageCandidatePaths(candidates []UnstageCandidate) []string {
	paths := make([]string, 0, len(candidates))
	for _, candidate := range candidates {
		paths = append(paths, candidate.Path)
	}
	return paths
}

// DetectUnstageCandidates reports which of paths gen-commit-msg would unstage.
// A nil paths list means "the current staged set". Nothing is mutated and the
// process working directory is restored before returning.
//
// This is the single detector behind both the --dry-run plan and the real
// unstage step, so external callers and generation can never disagree about
// which paths are excluded.
func DetectUnstageCandidates(dir string, paths []string) ([]UnstageCandidate, error) {
	return detectUnstageCandidates(dir, paths)
}

func detectUnstageCandidates(dir string, paths []string) ([]UnstageCandidate, error) {
	dir = filepath.Clean(dir)

	repoRoot, err := git.ShowToplevel(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve git toplevel: %w", err)
	}
	if repoRoot == "" {
		return nil, fmt.Errorf("empty git toplevel")
	}

	// File type detection and submodule matching both work on repo-relative
	// paths, so evaluate everything from the root (same as detectAndUnstage).
	origDir, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("getwd: %w", err)
	}
	if err := os.Chdir(repoRoot); err != nil {
		return nil, fmt.Errorf("chdir %s: %w", repoRoot, err)
	}
	defer os.Chdir(origDir)

	if paths == nil {
		paths, err = git.GetStagedFiles(repoRoot)
		if err != nil {
			return nil, fmt.Errorf("failed to list staged files: %w", err)
		}
	}
	if len(paths) == 0 {
		return nil, nil
	}

	subModuleDirs := submodule.DetectSubModules(paths)
	candidates := make([]UnstageCandidate, 0, len(paths))
	for _, path := range paths {
		if subModuleDir := submoduleDirFor(path, subModuleDirs); subModuleDir != "" {
			candidates = append(candidates, UnstageCandidate{Path: path, Kind: UnstageKindSubmodule, SubmoduleDir: subModuleDir})
			continue
		}
		info, err := os.Stat(path)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				continue
			}
			return nil, fmt.Errorf("stat %s: %w", path, err)
		}
		if info.IsDir() {
			continue
		}
		desc, isBin, err := detect.DetectFileType(path)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				continue
			}
			return nil, fmt.Errorf("detect %s: %w", path, err)
		}
		if isBin {
			candidates = append(candidates, UnstageCandidate{Path: path, Kind: UnstageKindBinary, Desc: desc})
		}
	}

	return candidates, nil
}

// submoduleDirFor returns the submodule directory containing path, or "".
func submoduleDirFor(path string, subModuleDirs []string) string {
	for _, subModuleDir := range subModuleDirs {
		if strings.HasPrefix(path, subModuleDir+string(filepath.Separator)) || path == subModuleDir {
			return subModuleDir
		}
	}
	return ""
}

func detectAndUnstage(dir string, logger Logger) error {
	candidates, err := detectUnstageCandidates(dir, nil)
	if err != nil {
		return err
	}
	if len(candidates) == 0 {
		return nil
	}

	// Report and unstage binaries first, then submodule paths, matching the
	// order of the summary printed below.
	var binaries []UnstageCandidate
	binaryPaths := make([]string, 0, len(candidates))
	subModuleDirs := make([]string, 0, len(candidates))
	subModulePaths := make([]string, 0, len(candidates))
	seenSubModuleDir := make(map[string]bool, len(candidates))
	for _, candidate := range candidates {
		switch candidate.Kind {
		case UnstageKindBinary:
			binaries = append(binaries, candidate)
			binaryPaths = append(binaryPaths, candidate.Path)
		case UnstageKindSubmodule:
			subModulePaths = append(subModulePaths, candidate.Path)
			if !seenSubModuleDir[candidate.SubmoduleDir] {
				seenSubModuleDir[candidate.SubmoduleDir] = true
				subModuleDirs = append(subModuleDirs, candidate.SubmoduleDir)
			}
		}
	}
	toUnstage := make([]string, 0, len(candidates))
	toUnstage = append(toUnstage, binaryPaths...)
	toUnstage = append(toUnstage, subModulePaths...)

	if len(binaries) > 0 {
		fmt.Fprintf(os.Stderr, "\nAuto-unstaged binary files:\n")
		for _, candidate := range binaries {
			if candidate.Desc != "" {
				fmt.Fprintf(os.Stderr, "  %s (%s)\n", candidate.Path, candidate.Desc)
			} else {
				fmt.Fprintf(os.Stderr, "  %s\n", candidate.Path)
			}
		}
	}
	if len(subModuleDirs) > 0 {
		fmt.Fprintf(os.Stderr, "\nAuto-unstaged submodule directories:\n")
		for _, subModuleDir := range subModuleDirs {
			fmt.Fprintf(os.Stderr, "  %s/\n", subModuleDir)
		}
	}

	fmt.Fprintf(os.Stderr, "\nTo include these files back: use `git add <file> && git commit --amend --no-edit`\n")

	logger.Log("Unstaging binary/submodule entries...")
	if err := gitwrite.RestoreStaged(dir, toUnstage...); err != nil {
		return err
	}

	return nil
}
