package markdown

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

type MetaEntry struct {
	Name    string
	Content string
}

type MetaMap []MetaEntry

func (m MetaMap) ToMap() map[string]string {
	result := make(map[string]string, len(m))
	for _, e := range m {
		result[e.Name] = e.Content
	}
	return result
}

type Document struct {
	Meta    MetaMap
	Content string
}

func ParseWithMeta(path string) (*Document, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)

	var fmLines []string
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}
		if line == "---" {
			if len(fmLines) == 0 {
				continue
			}
			break
		}
		fmLines = append(fmLines, line)
	}

	if len(fmLines) == 0 {
		return nil, errors.New("frontmatter not found")
	}

	var root yaml.Node
	if err := yaml.Unmarshal([]byte(strings.Join(fmLines, "\n")), &root); err != nil {
		return nil, fmt.Errorf("parse frontmatter yaml: %w", err)
	}

	var meta MetaMap
	if root.Kind == yaml.DocumentNode && len(root.Content) > 0 {
		doc := root.Content[0]
		if doc.Kind == yaml.MappingNode {
			for i := 0; i < len(doc.Content); i += 2 {
				key := doc.Content[i].Value
				val := nodeValue(doc.Content[i+1])
				meta = append(meta, MetaEntry{Name: key, Content: val})
			}
		}
	}

	var bodyLines []string
	for scanner.Scan() {
		bodyLines = append(bodyLines, scanner.Text())
	}

	return &Document{
		Meta:    meta,
		Content: strings.TrimSpace(strings.Join(bodyLines, "\n")),
	}, nil
}

func nodeValue(n *yaml.Node) string {
	switch n.Kind {
	case yaml.ScalarNode:
		return n.Value
	default:
		out, _ := yaml.Marshal(n)
		return string(out)
	}
}
