package hooks

import (
	"fmt"

	"github.com/xhd2015/agent-pro/markdown"
)

func ParseTemplate(path string) (*Entry, error) {
	doc, err := markdown.ParseWithMeta(path)
	if err != nil {
		return nil, err
	}

	m := doc.Meta.ToMap()

	e := &Entry{
		Template:    doc.Content,
		Name:        m["name"],
		Description: m["description"],
	}

	if e.Name == "" {
		return nil, fmt.Errorf("frontmatter missing required field: name")
	}

	return e, nil
}
