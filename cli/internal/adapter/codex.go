package adapter

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"

	"github.com/pelletier/go-toml/v2"
)

type codexRenderer struct{}

const codexConfigSchemaURL = "https://developers.openai.com/codex/config-schema.json"

func (codexRenderer) target() string {
	return "codex"
}

func (codexRenderer) manifestDirectory() string {
	return ".codex"
}

func (codexRenderer) unsupportedCategories() []string {
	return []string{"prompts", "rules"}
}

func (codexRenderer) render(base string) (map[string][]byte, error) {
	out := make(map[string][]byte)
	if err := codexInstructions(base, out); err != nil {
		return nil, err
	}
	if err := copyDirectory(filepath.Join(base, "instructions", "codex-project"), ".", out); err != nil {
		return nil, err
	}
	if err := codexAgents(base, out); err != nil {
		return nil, err
	}
	if err := codexConfig(base, out); err != nil {
		return nil, err
	}
	if err := copyDirectory(filepath.Join(base, "permissions", "codex-rules"), ".codex/rules", out); err != nil {
		return nil, err
	}
	// Codex discovers repository skills directly from .agents/skills. Validate
	// the canonical skill tree without projecting a duplicate registration.
	if err := skills(base); err != nil {
		return nil, err
	}
	if err := codexLegacySkills(base, out); err != nil {
		return nil, err
	}
	return out, nil
}

type codexLegacySkillsManifest struct {
	Files []string `json:"files"`
}

func codexLegacySkills(base string, out map[string][]byte) error {
	data, err := readSourceFile(filepath.Join(base, "settings", "codex-legacy-skills.json"))
	if errors.Is(err, fs.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	var manifest codexLegacySkillsManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return fmt.Errorf("settings/codex-legacy-skills.json: %w", err)
	}
	seen := map[string]struct{}{}
	for _, relative := range manifest.Files {
		clean := filepath.ToSlash(filepath.Clean(filepath.FromSlash(relative)))
		if clean == "." || filepath.IsAbs(relative) || strings.HasPrefix(clean, "../") {
			return fmt.Errorf("settings/codex-legacy-skills.json: invalid path %q", relative)
		}
		if _, exists := seen[clean]; exists {
			return fmt.Errorf("settings/codex-legacy-skills.json: duplicate path %q", relative)
		}
		seen[clean] = struct{}{}
		content, err := readSourceFile(filepath.Join(base, "skills", filepath.FromSlash(clean)))
		if err != nil {
			return fmt.Errorf("legacy Codex skill %s: %w", clean, err)
		}
		destination := filepath.ToSlash(filepath.Join(".codex", "skills", filepath.FromSlash(clean)))
		if _, exists := out[destination]; exists {
			return fmt.Errorf("duplicate projected output path %q", destination)
		}
		out[destination] = content
	}
	return nil
}

func codexInstructions(base string, out map[string][]byte) error {
	files, err := markdownFiles(filepath.Join(base, "instructions"))
	if err != nil {
		return err
	}
	var root []byte
	var sections []string
	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			return err
		}
		switch filepath.Base(file) {
		case "AGENTS.md":
			root = data
		case "copilot-instructions.md", "CLAUDE.md":
			continue
		default:
			sections = append(sections, strings.TrimSpace(string(data)))
		}
	}
	if len(root) > 0 && len(sections) == 0 {
		out["AGENTS.md"] = root
		return nil
	}
	if len(root) > 0 {
		sections = append([]string{strings.TrimSpace(string(root))}, sections...)
	}
	if len(sections) > 0 {
		out["AGENTS.md"] = []byte(strings.Join(sections, "\n\n") + "\n")
	}
	return nil
}

func codexAgents(base string, out map[string][]byte) error {
	files, err := markdownFiles(filepath.Join(base, "agents"))
	if err != nil {
		return err
	}
	consumedSidecars := map[string]bool{}
	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			return err
		}
		front, body, err := frontMatter(data)
		if err != nil {
			return fmt.Errorf("agent %s: %w", file, err)
		}
		description, ok := front["description"].(string)
		if !ok || strings.TrimSpace(description) == "" {
			return fmt.Errorf("agent %s: front matter requires description", file)
		}
		name := strings.TrimSuffix(filepath.Base(file), ".md")
		if configured, ok := front["name"].(string); ok && strings.TrimSpace(configured) != "" {
			name = configured
		}
		config := map[string]any{}
		destinationName := strings.TrimSuffix(filepath.Base(file), ".md") + ".toml"
		rawPath := filepath.Join(base, "settings", "codex-agents", destinationName)
		if info, statErr := os.Lstat(rawPath); statErr == nil && info.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("refusing symlinked Codex agent sidecar %s", rawPath)
		} else if statErr != nil && !errors.Is(statErr, fs.ErrNotExist) {
			return statErr
		}
		raw, rawErr := os.ReadFile(rawPath)
		if rawErr == nil {
			consumedSidecars[destinationName] = true
			if err := toml.Unmarshal(raw, &config); err != nil {
				return fmt.Errorf("settings/codex-agents/%s: %w", destinationName, err)
			}
		} else if !errors.Is(rawErr, fs.ErrNotExist) {
			return rawErr
		}
		config["name"] = name
		config["description"] = description
		config["developer_instructions"] = strings.TrimSpace(string(body))
		if len(raw) == 0 {
			content := fmt.Sprintf("name = %s\ndescription = %s\ndeveloper_instructions = %s\n",
				strconv.Quote(name), strconv.Quote(description), strconv.Quote(strings.TrimSpace(string(body))))
			out[filepath.ToSlash(filepath.Join(".codex/agents", destinationName))] = []byte(content)
			continue
		}
		content, err := marshalTOMLPreservingRaw(config, raw)
		if err != nil {
			return fmt.Errorf("agent %s: %w", file, err)
		}
		out[filepath.ToSlash(filepath.Join(".codex/agents", destinationName))] = content
	}
	sidecarDir := filepath.Join(base, "settings", "codex-agents")
	entries, err := os.ReadDir(sidecarDir)
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return err
	}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".toml") {
			continue
		}
		if !consumedSidecars[entry.Name()] {
			return fmt.Errorf("settings/codex-agents/%s has no matching canonical agents/%s.md", entry.Name(), strings.TrimSuffix(entry.Name(), ".toml"))
		}
	}
	return nil
}

func codexConfig(base string, out map[string][]byte) error {
	config := map[string]any{}
	raw, err := readSourceFile(filepath.Join(base, "settings", "codex.raw.toml"))
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return err
	}
	for _, category := range []string{"settings", "guardrails", "hooks", "memories", "permissions", "plugins", "profiles"} {
		data, err := readSourceFile(filepath.Join(base, category, "codex.toml"))
		if errors.Is(err, fs.ErrNotExist) {
			continue
		}
		if err != nil {
			return err
		}
		var fragment map[string]any
		if err := toml.Unmarshal(data, &fragment); err != nil {
			return fmt.Errorf("%s/codex.toml: %w", category, err)
		}
		for key, value := range fragment {
			config[key] = value
		}
	}
	if err := mergeCodexMCP(base, config); err != nil {
		return err
	}
	if len(config) == 0 {
		if len(raw) > 0 {
			var original map[string]any
			if err := toml.Unmarshal(raw, &original); err == nil && len(original) == 0 {
				out[".codex/config.toml"] = raw
			}
		}
		return nil
	}
	if len(raw) > 0 {
		var original map[string]any
		if err := toml.Unmarshal(raw, &original); err == nil && reflect.DeepEqual(normalizeValue(original), normalizeValue(config)) {
			out[".codex/config.toml"] = raw
			return nil
		}
	}
	content, err := marshalTOMLPreservingRaw(config, raw)
	if err != nil {
		return err
	}
	if !bytes.HasPrefix(content, []byte("#:schema ")) {
		content = append([]byte("#:schema "+codexConfigSchemaURL+"\n\n"), content...)
	}
	out[".codex/config.toml"] = content
	return nil
}

func mergeCodexMCP(base string, config map[string]any) error {
	data, err := readSourceFile(filepath.Join(base, "tools", "mcp.json"))
	if errors.Is(err, fs.ErrNotExist) {
		delete(config, "mcp_servers")
		return nil
	}
	if err != nil {
		return err
	}
	var portable struct {
		Servers map[string]map[string]any `json:"mcpServers"`
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	if err := decoder.Decode(&portable); err != nil || portable.Servers == nil {
		return errors.New("tools/mcp.json must contain an mcpServers object")
	}
	if len(portable.Servers) == 0 {
		delete(config, "mcp_servers")
		return nil
	}
	servers := make(map[string]any, len(portable.Servers))
	for name, source := range portable.Servers {
		source = normalizeStringMap(source)
		server := map[string]any{}
		if extension, ok := stringMap(source["codex"]); ok {
			for key, value := range extension {
				server[key] = value
			}
		}
		typeName, _ := source["type"].(string)
		switch typeName {
		case "", "stdio", "local":
			command, ok := source["command"].(string)
			if !ok || strings.TrimSpace(command) == "" {
				return fmt.Errorf("MCP server %q requires command for Codex stdio transport", name)
			}
			server["command"] = command
			copyPortableField(source, server, "args", "env", "cwd")
		case "http", "streamable-http":
			url, ok := source["url"].(string)
			if !ok || strings.TrimSpace(url) == "" {
				return fmt.Errorf("MCP server %q requires url for Codex HTTP transport", name)
			}
			server["url"] = url
			if headers, exists := source["headers"]; exists {
				server["http_headers"] = headers
			}
		default:
			return fmt.Errorf("MCP server %q uses unsupported Codex transport %q", name, typeName)
		}
		servers[name] = server
	}
	config["mcp_servers"] = servers
	return nil
}

func copyPortableField(source, destination map[string]any, keys ...string) {
	for _, key := range keys {
		if value, ok := source[key]; ok {
			destination[key] = value
		}
	}
}

func marshalTOMLPreservingRaw(config map[string]any, raw []byte) ([]byte, error) {
	if len(raw) > 0 {
		var original map[string]any
		if err := toml.Unmarshal(raw, &original); err == nil && reflect.DeepEqual(normalizeValue(original), normalizeValue(config)) {
			return raw, nil
		}
	}
	data, err := toml.Marshal(config)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func stringMap(value any) (map[string]any, bool) {
	result, ok := value.(map[string]any)
	return result, ok
}

func normalizeStringMap(value map[string]any) map[string]any {
	normalized, _ := normalizeValue(value).(map[string]any)
	return normalized
}

func normalizeValue(value any) any {
	switch typed := value.(type) {
	case json.Number:
		if integer, err := typed.Int64(); err == nil {
			return integer
		}
		if decimal, err := typed.Float64(); err == nil {
			return decimal
		}
		return typed.String()
	case map[string]any:
		out := make(map[string]any, len(typed))
		for key, item := range typed {
			out[key] = normalizeValue(item)
		}
		return out
	case []any:
		out := make([]any, len(typed))
		for index, item := range typed {
			out[index] = normalizeValue(item)
		}
		return out
	case []string:
		out := make([]any, len(typed))
		for index, item := range typed {
			out[index] = item
		}
		return out
	default:
		return value
	}
}
