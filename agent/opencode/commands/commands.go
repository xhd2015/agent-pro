package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/xhd2015/agent-pro/agent/opencode/config"
	"github.com/xhd2015/agent-pro/markdown"
)

type Command struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Agent       string `json:"agent,omitempty"`
	Model       string `json:"model,omitempty"`
	Subtask     bool   `json:"subtask,omitempty"`
	Template    string `json:"template,omitempty"`
	Source      string `json:"source,omitempty"`
	Path        string `json:"path,omitempty"`
}

func List(opencodeDir string) ([]Command, error) {
	var all []Command

	jsonCmds := listFromConfigData(opencodeDir)
	all = append(all, jsonCmds...)

	mdCmds, err := listFromMarkdown(opencodeDir)
	if err != nil {
		return nil, fmt.Errorf("list markdown commands: %w", err)
	}

	seen := make(map[string]bool)
	for _, c := range all {
		seen[c.Name] = true
	}
	for _, c := range mdCmds {
		if !seen[c.Name] {
			all = append(all, c)
			seen[c.Name] = true
		}
	}

	sort.Slice(all, func(i, j int) bool { return all[i].Name < all[j].Name })
	return all, nil
}

func GetByNames(data config.Data, names map[string]bool) []Command {
	cmdMap := getCommandMap(data)
	if cmdMap == nil {
		return nil
	}

	var result []Command
	for name, val := range cmdMap {
		if !names[name] {
			continue
		}
		c := Command{Name: name}
		if obj, ok := val.(map[string]interface{}); ok {
			fillFromConfigObj(&c, obj)
		}
		result = append(result, c)
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Name < result[j].Name })
	return result
}

func Merge(data config.Data, commands []Command) (added int, addedNames []string, existing []string) {
	cmdMap := ensureCommandMap(data)

	for _, c := range commands {
		if _, exists := cmdMap[c.Name]; !exists {
			cmdMap[c.Name] = map[string]interface{}{
				"template":    c.Template,
				"description": c.Description,
			}
			added++
			addedNames = append(addedNames, c.Name)
		} else {
			existing = append(existing, c.Name)
		}
	}

	return
}

func RemoveFromMap(cmdMap map[string]interface{}, commands []Command) (removed int, removedNames []string) {
	for _, c := range commands {
		if _, exists := cmdMap[c.Name]; exists {
			delete(cmdMap, c.Name)
			removed++
			removedNames = append(removedNames, c.Name)
		}
	}
	return
}

func ParseTemplate(path string) (*Command, error) {
	doc, err := markdown.ParseWithMeta(path)
	if err != nil {
		return nil, err
	}

	m := doc.Meta.ToMap()

	c := &Command{
		Template:    doc.Content,
		Name:        m["name"],
		Description: m["description"],
		Agent:       m["agent"],
		Model:       m["model"],
	}

	if subtask, ok := m["subtask"]; ok {
		c.Subtask = subtask == "true"
	}

	if c.Name == "" {
		return nil, fmt.Errorf("frontmatter missing required field: name")
	}

	return c, nil
}

func ensureCommandMap(data config.Data) map[string]interface{} {
	cmd, ok := data["command"]
	if !ok {
		cmd = map[string]interface{}{}
		data["command"] = cmd
	}
	cmdMap, ok := cmd.(map[string]interface{})
	if !ok {
		cmdMap = map[string]interface{}{}
		data["command"] = cmdMap
	}
	return cmdMap
}

func getCommandMap(data config.Data) map[string]interface{} {
	cmd, ok := data["command"]
	if !ok {
		return nil
	}
	cmdMap, ok := cmd.(map[string]interface{})
	if !ok {
		return nil
	}
	return cmdMap
}

func listFromConfigData(opencodeDir string) []Command {
	cfg, err := config.Read(filepath.Dir(opencodeDir))
	if err != nil {
		return nil
	}

	cmdMap := getCommandMap(cfg.Data)
	if cmdMap == nil {
		return nil
	}

	var result []Command
	for name, val := range cmdMap {
		c := Command{Name: name, Source: "config"}
		if obj, ok := val.(map[string]interface{}); ok {
			fillFromConfigObj(&c, obj)
		}
		result = append(result, c)
	}
	return result
}

func listFromMarkdown(opencodeDir string) ([]Command, error) {
	var result []Command

	for _, sub := range []string{"command", "commands"} {
		subDir := filepath.Join(opencodeDir, sub)
		err := filepath.WalkDir(subDir, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return nil
			}
			if d.IsDir() || !strings.HasSuffix(d.Name(), ".md") {
				return nil
			}
			cmd, parseErr := ParseTemplate(path)
			if parseErr != nil {
				return nil
			}
			cmd.Source = "markdown"
			cmd.Path = path
			result = append(result, *cmd)
			return nil
		})
		if err != nil && !os.IsNotExist(err) {
			return nil, fmt.Errorf("walk %s: %w", subDir, err)
		}
	}
	return result, nil
}

func fillFromConfigObj(c *Command, obj map[string]interface{}) {
	if desc, ok := obj["description"].(string); ok {
		c.Description = desc
	}
	if agent, ok := obj["agent"].(string); ok {
		c.Agent = agent
	}
	if model, ok := obj["model"].(string); ok {
		c.Model = model
	}
	if template, ok := obj["template"].(string); ok {
		c.Template = template
	}
	if subtask, ok := obj["subtask"].(bool); ok {
		c.Subtask = subtask
	}
}
