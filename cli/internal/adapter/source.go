package adapter

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

var allCategoryNames = []string{
	"agents",
	"guardrails",
	"hooks",
	"instructions",
	"memories",
	"permissions",
	"plugins",
	"profiles",
	"prompts",
	"rules",
	"settings",
	"skills",
	"tools",
}

func allCategories() []string {
	return append([]string(nil), allCategoryNames...)
}

func instructions(base, destination string, out map[string][]byte) error {
	files, err := markdownFiles(filepath.Join(base, "instructions"))
	if err != nil {
		return err
	}
	var sections []string
	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			return err
		}
		sections = append(sections, strings.TrimSpace(string(data)))
	}
	if len(sections) > 0 {
		out[destination] = []byte(strings.Join(sections, "\n\n") + "\n")
	}
	return nil
}

func rules(base, destination, suffix string, out map[string][]byte) error {
	files, err := markdownFiles(filepath.Join(base, "rules"))
	if err != nil {
		return err
	}
	applied := map[string]string{}
	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			return err
		}
		front, _, err := frontMatter(data)
		if err != nil {
			return fmt.Errorf("rule %s: %w", file, err)
		}
		applyTo, ok := front["applyTo"].(string)
		if !ok || strings.TrimSpace(applyTo) == "" {
			return fmt.Errorf("rule %s: front matter requires applyTo", file)
		}
		if source, exists := applied[applyTo]; exists {
			return fmt.Errorf("rule %s and %s both target applyTo %q", filepath.Base(source), filepath.Base(file), applyTo)
		}
		applied[applyTo] = file
		out[filepath.ToSlash(filepath.Join(destination, strings.TrimSuffix(filepath.Base(file), ".md")+suffix))] = data
	}
	return nil
}

func claudeRules(base string, out map[string][]byte) error {
	files, err := markdownFiles(filepath.Join(base, "rules"))
	if err != nil {
		return err
	}
	applied := map[string]string{}
	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			return err
		}
		front, body, err := frontMatter(data)
		if err != nil {
			return fmt.Errorf("rule %s: %w", file, err)
		}
		applyTo, ok := front["applyTo"].(string)
		if !ok || strings.TrimSpace(applyTo) == "" {
			return fmt.Errorf("rule %s: front matter requires applyTo", file)
		}
		paths := splitGlobPatterns(applyTo)
		if len(paths) == 0 {
			return fmt.Errorf("rule %s: front matter requires applyTo", file)
		}
		for _, path := range paths {
			if source, exists := applied[path]; exists {
				return fmt.Errorf("rule %s and %s both target path pattern %q", filepath.Base(source), filepath.Base(file), path)
			}
			applied[path] = file
		}
		delete(front, "applyTo")
		front["paths"] = paths
		encoded, err := yaml.Marshal(front)
		if err != nil {
			return fmt.Errorf("rule %s: %w", file, err)
		}
		output := append([]byte("---\n"), encoded...)
		output = append(output, []byte("---\n")...)
		output = append(output, body...)
		out[filepath.ToSlash(filepath.Join(".claude/rules", filepath.Base(file)))] = output
	}
	return nil
}

func agents(base, destination, suffix string, out map[string][]byte) error {
	files, err := markdownFiles(filepath.Join(base, "agents"))
	if err != nil {
		return err
	}
	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			return err
		}
		front, _, err := frontMatter(data)
		if err != nil {
			return fmt.Errorf("agent %s: %w", file, err)
		}
		description, ok := front["description"].(string)
		if !ok || strings.TrimSpace(description) == "" {
			return fmt.Errorf("agent %s: front matter requires description", file)
		}
		out[filepath.ToSlash(filepath.Join(destination, strings.TrimSuffix(filepath.Base(file), ".md")+suffix))] = data
	}
	return nil
}

func copyJSON(base, source, destination string, out map[string][]byte) error {
	dir := filepath.Join(base, source)
	entries, err := os.ReadDir(dir)
	if errors.Is(err, fs.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(dir, entry.Name()))
		if err != nil {
			return err
		}
		if !json.Valid(data) {
			return fmt.Errorf("%s is not valid JSON", filepath.Join(source, entry.Name()))
		}
		out[filepath.ToSlash(filepath.Join(destination, entry.Name()))] = data
	}
	return nil
}

func mcpJSON(base, destination string, out map[string][]byte) error {
	path := filepath.Join(base, "tools", "mcp.json")
	data, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	var config struct {
		Servers map[string]json.RawMessage `json:"mcpServers"`
	}
	if err := json.Unmarshal(data, &config); err != nil || config.Servers == nil {
		return errors.New("tools/mcp.json must contain an mcpServers object")
	}
	for name, raw := range config.Servers {
		if strings.TrimSpace(name) == "" {
			return errors.New("tools/mcp.json must not contain an empty MCP server name")
		}
		var server map[string]json.RawMessage
		if err := json.Unmarshal(raw, &server); err != nil || server == nil {
			return fmt.Errorf("MCP server %q must be an object", name)
		}
	}
	out[destination] = data
	return nil
}

func skills(base string) error {
	dir := filepath.Join(base, "skills")
	entries, err := os.ReadDir(dir)
	if errors.Is(err, fs.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		data, err := os.ReadFile(filepath.Join(dir, entry.Name(), "SKILL.md"))
		if err != nil {
			return fmt.Errorf("skill %s must contain SKILL.md", entry.Name())
		}
		front, _, err := frontMatter(data)
		name, nameOK := front["name"].(string)
		description, descriptionOK := front["description"].(string)
		if err != nil || !nameOK || strings.TrimSpace(name) == "" ||
			!descriptionOK || strings.TrimSpace(description) == "" {
			return fmt.Errorf("skill %s requires name and description front matter", entry.Name())
		}
	}
	return nil
}

func copySkills(base, destination string, out map[string][]byte) error {
	if err := skills(base); err != nil {
		return err
	}
	source := filepath.Join(base, "skills")
	if _, err := os.Stat(source); errors.Is(err, fs.ErrNotExist) {
		return nil
	} else if err != nil {
		return err
	}
	entries, err := os.ReadDir(source)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		skillRoot := filepath.Join(source, entry.Name())
		err := filepath.WalkDir(skillRoot, func(path string, entry fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if entry.IsDir() {
				return nil
			}
			relative, err := filepath.Rel(source, path)
			if err != nil {
				return err
			}
			data, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			out[filepath.ToSlash(filepath.Join(destination, relative))] = data
			return nil
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func claudeHooks(base string, out map[string][]byte) error {
	dir := filepath.Join(base, "hooks")
	entries, err := os.ReadDir(dir)
	if errors.Is(err, fs.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	hooks := make(map[string]json.RawMessage)
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(dir, entry.Name()))
		if err != nil {
			return err
		}
		var config struct {
			Hooks map[string]json.RawMessage `json:"hooks"`
		}
		if err := json.Unmarshal(data, &config); err != nil || config.Hooks == nil {
			return fmt.Errorf("hooks/%s must contain a hooks object for Claude", entry.Name())
		}
		for event, definition := range config.Hooks {
			if _, exists := hooks[event]; exists {
				return fmt.Errorf("Claude hook event %q is declared by more than one file", event)
			}
			hooks[event] = definition
		}
	}
	if len(hooks) > 0 {
		data, err := json.MarshalIndent(struct {
			Hooks map[string]json.RawMessage `json:"hooks"`
		}{Hooks: hooks}, "", "  ")
		if err != nil {
			return err
		}
		out[".claude/settings.json"] = append(data, '\n')
	}
	return nil
}

func sortedKeys(values map[string]string) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func tomlStrings(values []string) string {
	quoted := make([]string, len(values))
	for i, value := range values {
		quoted[i] = strconv.Quote(value)
	}
	return "[" + strings.Join(quoted, ", ") + "]"
}

func tomlKey(key string) string {
	return strconv.Quote(key)
}

func markdownFiles(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var files []string
	for _, entry := range entries {
		if !entry.IsDir() && entry.Name() != "README.md" && strings.HasSuffix(entry.Name(), ".md") {
			files = append(files, filepath.Join(dir, entry.Name()))
		}
	}
	sort.Strings(files)
	return files, nil
}

func frontMatter(data []byte) (map[string]any, []byte, error) {
	text := string(data)
	if !strings.HasPrefix(text, "---\n") {
		return nil, nil, errors.New("must start with YAML front matter")
	}
	end := strings.Index(text[4:], "\n---\n")
	if end < 0 {
		return nil, nil, errors.New("front matter is not closed")
	}
	var front map[string]any
	if err := yaml.Unmarshal([]byte(text[4:4+end]), &front); err != nil {
		return nil, nil, err
	}
	return front, []byte(text[4+end+5:]), nil
}

func splitGlobPatterns(value string) []string {
	var patterns []string
	braceDepth := 0
	start := 0
	for index, char := range value {
		switch char {
		case '{':
			braceDepth++
		case '}':
			if braceDepth > 0 {
				braceDepth--
			}
		case ',':
			if braceDepth == 0 {
				if pattern := strings.TrimSpace(value[start:index]); pattern != "" {
					patterns = append(patterns, pattern)
				}
				start = index + 1
			}
		}
	}
	if pattern := strings.TrimSpace(value[start:]); pattern != "" {
		patterns = append(patterns, pattern)
	}
	return patterns
}
