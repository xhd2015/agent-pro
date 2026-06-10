package spec

import (
	"fmt"

	"github.com/xhd2015/agent-pro/agents/doctest/doc"
	"github.com/xhd2015/skills/install"
)

type entry struct {
	SkillName string
	FileName  string
}

var entries = map[string]entry{
	"doc-spec":  {SkillName: "doc-style-test-specification", FileName: "DOC_STYLE_TEST_SPECIFICATION.md"},
	"code-spec": {SkillName: "doc-style-test-code-specification", FileName: "DOC_STYLE_TEST_CODE_SPECIFICATION.md"},
}

func Content(name string) (string, error) {
	ent, ok := entries[name]
	if !ok {
		return "", fmt.Errorf("unknown skill: %s", name)
	}
	return doc.Content(ent.FileName)
}

func Install(name string, args []string) error {
	ent, ok := entries[name]
	if !ok {
		return fmt.Errorf("unknown skill: %s", name)
	}
	content, err := Content(name)
	if err != nil {
		return err
	}
	return install.HandleInstall(install.InstallOptions{
		SkillDirName: ent.SkillName,
		SkillContent: content,
		Usage:        "doctest skill " + name + " install",
	}, args)
}
