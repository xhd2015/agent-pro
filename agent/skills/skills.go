package skills

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// SkillInfo holds parsed metadata for a single skill.
type SkillInfo struct {
	Name        string
	Description string
	Path        string
}

// ListDir scans a directory for subdirectories containing a SKILL.md file
// with valid YAML frontmatter (name + description). Dot-prefixed dirs are skipped.
func ListDir(root string) ([]SkillInfo, error) {
	entries, err := os.ReadDir(root)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var skills []SkillInfo
	for _, entry := range entries {
		if !entry.IsDir() || strings.HasPrefix(entry.Name(), ".") {
			continue
		}
		skillPath := filepath.Join(root, entry.Name())
		mdPath := filepath.Join(skillPath, "SKILL.md")
		info, err := ParseSkillMD(mdPath)
		if err != nil {
			continue
		}
		if info == nil {
			continue
		}
		info.Path = skillPath
		skills = append(skills, *info)
	}
	return skills, nil
}

// ParseSkillMD reads a SKILL.md file and extracts YAML frontmatter name/description.
// Returns nil if the file has no valid frontmatter.
func ParseSkillMD(path string) (*SkillInfo, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	var frontmatter []string
	inFM := false
	lineCount := 0

	for scanner.Scan() {
		line := scanner.Text()
		lineCount++
		if lineCount == 1 {
			if strings.TrimSpace(line) == "---" {
				inFM = true
				continue
			}
			return nil, fmt.Errorf("no frontmatter")
		}
		if inFM {
			if strings.TrimSpace(line) == "---" {
				break
			}
			frontmatter = append(frontmatter, line)
		}
	}

	if len(frontmatter) == 0 {
		return nil, nil
	}

	var fm struct {
		Name        string `yaml:"name"`
		Description string `yaml:"description"`
	}
	if err := yaml.Unmarshal([]byte(strings.Join(frontmatter, "\n")), &fm); err != nil {
		return nil, fmt.Errorf("yaml parse: %w", err)
	}

	if fm.Name == "" {
		return nil, nil
	}

	return &SkillInfo{
		Name:        fm.Name,
		Description: fm.Description,
	}, nil
}
