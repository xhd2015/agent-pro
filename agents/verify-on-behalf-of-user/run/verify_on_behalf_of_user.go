package run

import (
	"embed"
	"fmt"
	"io/fs"
	"path"
	"strings"

	"github.com/xhd2015/skills/skillcmd"
)

//go:embed SKILL.md
var SkillFile string

//go:embed sandbox
//go:embed workflow
//go:embed transcript
var skillTreeFS embed.FS

//go:embed scripts/enter-sandbox.sh
var enterSandboxScript []byte

//go:embed scripts/sandbox-verify.sb
var sandboxVerifySB []byte

//go:embed templates/transcript.md
var transcriptTemplate []byte

// SkillTree returns the embedded topic tree (path → path/TOPIC.md).
func SkillTree() fs.FS {
	return skillTreeFS
}

// TopicContent loads a nested topic by slash-separated path (e.g. "sandbox").
func TopicContent(topicPath string) (string, error) {
	topicPath = strings.Trim(topicPath, "/")
	if topicPath == "" {
		return "", fmt.Errorf("empty topic path")
	}
	for _, s := range strings.Split(topicPath, "/") {
		if s == "" || s == "." || s == ".." {
			return "", fmt.Errorf("invalid topic path segment: %q", s)
		}
	}
	data, err := skillTreeFS.ReadFile(path.Join(topicPath, "TOPIC.md"))
	if err != nil {
		return "", fmt.Errorf("unknown topic: %s", topicPath)
	}
	return string(data), nil
}

// InstallFiles returns utility files and nested TOPIC.md topics bundled on install.
func InstallFiles() ([]skillcmd.InstallFile, error) {
	files := []skillcmd.InstallFile{
		{Path: "scripts/enter-sandbox.sh", Content: enterSandboxScript},
		{Path: "scripts/sandbox-verify.sb", Content: sandboxVerifySB},
		{Path: "templates/transcript.md", Content: transcriptTemplate},
	}
	nested, err := collectNestedTopicFiles(skillTreeFS)
	if err != nil {
		return nil, err
	}
	return append(files, nested...), nil
}

func collectNestedTopicFiles(treeFS fs.FS) ([]skillcmd.InstallFile, error) {
	var files []skillcmd.InstallFile
	err := fs.WalkDir(treeFS, ".", func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		p = path.Clean(p)
		if path.Base(p) != "TOPIC.md" {
			return nil
		}
		if p == "TOPIC.md" || p == "./TOPIC.md" {
			return nil
		}
		data, err := fs.ReadFile(treeFS, p)
		if err != nil {
			return err
		}
		files = append(files, skillcmd.InstallFile{Path: p, Content: data})
		return nil
	})
	return files, err
}