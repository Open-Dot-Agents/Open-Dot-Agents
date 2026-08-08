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
)

type codexRenderer struct{}

func (codexRenderer) target() string {
	return "codex"
}

func (codexRenderer) manifestDirectory() string {
	return ".codex"
}

func (codexRenderer) unsupportedCategories() []string {
	return append(append([]string{}, commonUnsupportedCategories...), "hooks", "rules")
}

func (codexRenderer) render(base string) (map[string][]byte, error) {
	out := make(map[string][]byte)
	if err := instructions(base, "AGENTS.md", out); err != nil {
		return nil, err
	}
	if err := codexAgents(base, out); err != nil {
		return nil, err
	}
	if err := codexMCP(base, out); err != nil {
		return nil, err
	}
	if err := copySkills(base, ".codex/skills", out); err != nil {
		return nil, err
	}
	return out, nil
}

func codexAgents(base string, out map[string][]byte) error {
	files, err := markdownFiles(filepath.Join(base, "agents"))
	if err != nil {
		return err
	}
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
		content := fmt.Sprintf("name = %s\ndescription = %s\ndeveloper_instructions = %s\n",
			strconv.Quote(name), strconv.Quote(description), strconv.Quote(strings.TrimSpace(string(body))))
		out[filepath.ToSlash(filepath.Join(".codex/agents", name+".toml"))] = []byte(content)
	}
	return nil
}

func codexMCP(base string, out map[string][]byte) error {
	data, err := os.ReadFile(filepath.Join(base, "tools", "mcp.json"))
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
	names := make([]string, 0, len(config.Servers))
	for name := range config.Servers {
		names = append(names, name)
	}
	sort.Strings(names)
	var sections []string
	for _, name := range names {
		var server struct {
			Type    string            `json:"type"`
			Command string            `json:"command"`
			Args    []string          `json:"args"`
			Env     map[string]string `json:"env"`
			URL     string            `json:"url"`
			Headers map[string]string `json:"headers"`
		}
		if err := json.Unmarshal(config.Servers[name], &server); err != nil {
			return fmt.Errorf("MCP server %q is not an object: %w", name, err)
		}
		table := "[mcp_servers." + tomlKey(name) + "]"
		switch server.Type {
		case "stdio", "local", "":
			if strings.TrimSpace(server.Command) == "" {
				return fmt.Errorf("MCP server %q requires command for Codex stdio transport", name)
			}
			lines := []string{table, "command = " + strconv.Quote(server.Command)}
			if len(server.Args) > 0 {
				lines = append(lines, "args = "+tomlStrings(server.Args))
			}
			if len(server.Env) > 0 {
				lines = append(lines, "[mcp_servers."+tomlKey(name)+".env]")
				for _, key := range sortedKeys(server.Env) {
					lines = append(lines, tomlKey(key)+" = "+strconv.Quote(server.Env[key]))
				}
			}
			sections = append(sections, strings.Join(lines, "\n"))
		case "http", "streamable-http":
			if strings.TrimSpace(server.URL) == "" {
				return fmt.Errorf("MCP server %q requires url for Codex HTTP transport", name)
			}
			lines := []string{table, "url = " + strconv.Quote(server.URL)}
			if len(server.Headers) > 0 {
				lines = append(lines, "[mcp_servers."+tomlKey(name)+".http_headers]")
				for _, key := range sortedKeys(server.Headers) {
					lines = append(lines, tomlKey(key)+" = "+strconv.Quote(server.Headers[key]))
				}
			}
			sections = append(sections, strings.Join(lines, "\n"))
		default:
			return fmt.Errorf("MCP server %q uses unsupported Codex transport %q", name, server.Type)
		}
	}
	if len(sections) > 0 {
		out[".codex/config.toml"] = []byte(strings.Join(sections, "\n\n") + "\n")
	}
	return nil
}
