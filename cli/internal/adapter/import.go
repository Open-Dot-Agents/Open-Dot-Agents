package adapter

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

type importerFunc func(root string) (map[string][]byte, error)

var importers = map[string]importerFunc{}

func registerImporter(target string, importer importerFunc) {
	if _, exists := importers[target]; exists {
		panic(fmt.Sprintf("duplicate importer registration for %q", target))
	}
	importers[target] = importer
}

func importerFor(target string) (importerFunc, error) {
	importer, ok := importers[target]
	if !ok {
		return nil, fmt.Errorf("no importer for target %q", target)
	}
	return importer, nil
}

func init() {
	registerImporter("copilot", importCopilot)
	registerImporter("codex", importCodex)
	registerImporter("claude", importClaude)
}

func addImportOutput(outputs map[string][]byte, path string, data []byte) error {
	if _, exists := outputs[path]; exists {
		return fmt.Errorf("import generated duplicate output path %q", path)
	}
	outputs[path] = data
	return nil
}

func importCopilot(root string) (map[string][]byte, error) {
	out := make(map[string][]byte)

	if err := importCopilotInstructions(root, out); err != nil {
		return nil, err
	}
	if err := importCopilotRules(root, out); err != nil {
		return nil, err
	}
	if err := importCopilotAgents(root, out); err != nil {
		return nil, err
	}
	if err := importCopilotHooks(root, out); err != nil {
		return nil, err
	}
	if err := importCopilotTools(root, out); err != nil {
		return nil, err
	}
	return out, nil
}

func importCopilotInstructions(root string, out map[string][]byte) error {
	data, err := os.ReadFile(filepath.Join(root, ".github", "copilot-instructions.md"))
	if errors.Is(err, fs.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	return addImportOutput(out, filepath.ToSlash(filepath.Join(".agents", "instructions", "copilot-instructions.md")), data)
}

func importCopilotRules(root string, out map[string][]byte) error {
	rules, err := collectFilesWithSuffix(filepath.Join(root, ".github", "instructions"), ".instructions.md")
	if err != nil {
		return err
	}
	for _, rule := range rules {
		data, err := os.ReadFile(rule)
		if err != nil {
			return err
		}
		name := strings.TrimSuffix(filepath.Base(rule), ".instructions.md")
		content := fmt.Sprintf("---\napplyTo: \"**/*\"\n---\n%s", normalizeMarkdownBody(data))
		if err := addImportOutput(out, filepath.ToSlash(filepath.Join(".agents", "rules", name+".md")), []byte(content)); err != nil {
			return err
		}
	}
	return nil
}

func importCopilotAgents(root string, out map[string][]byte) error {
	agents, err := collectFilesWithSuffix(filepath.Join(root, ".github", "agents"), ".agent.md")
	if err != nil {
		return err
	}
	for _, path := range agents {
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		name := strings.TrimSuffix(filepath.Base(path), ".agent.md")
		content := fmt.Sprintf("---\ndescription: %s\n---\n%s", name, normalizeMarkdownBody(data))
		if err := addImportOutput(out, filepath.ToSlash(filepath.Join(".agents", "agents", name+".md")), []byte(content)); err != nil {
			return err
		}
	}
	return nil
}

func importCopilotHooks(root string, out map[string][]byte) error {
	paths, err := collectFilesWithSuffix(filepath.Join(root, ".github", "hooks"), ".json")
	if err != nil {
		return err
	}
	for _, path := range paths {
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if !json.Valid(data) {
			return fmt.Errorf("%s is not valid JSON", filepath.ToSlash(filepath.Clean(path)))
		}
		if err := addImportOutput(out, filepath.ToSlash(filepath.Join(".agents", "hooks", filepath.Base(path))), data); err != nil {
			return err
		}
	}
	return nil
}

func importCopilotTools(root string, out map[string][]byte) error {
	data, err := readMCPFile(root, filepath.Join(".github", "mcp.json"))
	if errors.Is(err, fs.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	return addImportOutput(out, filepath.ToSlash(filepath.Join(".agents", "tools", "mcp.json")), data)
}

func importCodex(root string) (map[string][]byte, error) {
	out := make(map[string][]byte)

	data, err := os.ReadFile(filepath.Join(root, "AGENTS.md"))
	if err == nil {
		if err := addImportOutput(out, filepath.ToSlash(filepath.Join(".agents", "instructions", "AGENTS.md")), data); err != nil {
			return nil, err
		}
	} else if !errors.Is(err, fs.ErrNotExist) {
		return nil, err
	}

	agentPaths, err := collectFilesWithSuffix(filepath.Join(root, ".codex", "agents"), ".toml")
	if err != nil {
		return nil, err
	}
	for _, path := range agentPaths {
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		info, err := parseCodexAgentTOML(data)
		if err != nil {
			return nil, err
		}
		name := strings.TrimSuffix(filepath.Base(path), ".toml")
		if info.Name != "" {
			name = info.Name
		}
		content := fmt.Sprintf(
			"---\ndescription: %s\n---\n%s",
			strings.TrimSpace(info.Description),
			normalizeBody(info.Instructions),
		)
		if err := addImportOutput(out, filepath.ToSlash(filepath.Join(".agents", "agents", name+".md")), []byte(content)); err != nil {
			return nil, err
		}
	}

	mcpData, err := readCodexConfig(root)
	if !errors.Is(err, fs.ErrNotExist) {
		if err != nil {
			return nil, err
		}
		mcpJSON, err := parseCodexConfigTOML(mcpData)
		if err != nil {
			return nil, err
		}
		encoded, err := json.Marshal(mcpJSON)
		if err != nil {
			return nil, err
		}
		if err := addImportOutput(out, filepath.ToSlash(filepath.Join(".agents", "tools", "mcp.json")), append(encoded, '\n')); err != nil {
			return nil, err
		}
	}

	return out, nil
}

type codexAgentImport struct {
	Name         string
	Description  string
	Instructions string
}

func importClaude(root string) (map[string][]byte, error) {
	out := make(map[string][]byte)

	data, err := os.ReadFile(filepath.Join(root, "CLAUDE.md"))
	if err == nil {
		if err := addImportOutput(out, filepath.ToSlash(filepath.Join(".agents", "instructions", "CLAUDE.md")), data); err != nil {
			return nil, err
		}
	} else if !errors.Is(err, fs.ErrNotExist) {
		return nil, err
	}

	rulePaths, err := collectFilesWithSuffix(filepath.Join(root, ".claude", "rules"), ".md")
	if err != nil {
		return nil, err
	}
	for _, path := range rulePaths {
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		front, body, err := frontMatter(data)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", filepath.ToSlash(filepath.Clean(path)), err)
		}
		paths := stringSlice(front["paths"])
		if len(paths) == 0 {
			paths = stringSlice(front["applyTo"])
		}
		if len(paths) == 0 {
			return nil, fmt.Errorf("rule %s: front matter must include non-empty paths", filepath.ToSlash(filepath.Clean(path)))
		}
		name := strings.TrimSuffix(filepath.Base(path), ".md")
		content := fmt.Sprintf("---\napplyTo: \"%s\"\n---\n%s", strings.Join(paths, ", "), normalizeBody(string(body)))
		if err := addImportOutput(out, filepath.ToSlash(filepath.Join(".agents", "rules", name+".md")), []byte(content)); err != nil {
			return nil, err
		}
	}

	agentPaths, err := collectFilesWithSuffix(filepath.Join(root, ".claude", "agents"), ".md")
	if err != nil {
		return nil, err
	}
	for _, path := range agentPaths {
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		front, body, err := frontMatter(data)
		description := "Imported Claude agent"
		contentBody := data
		if err == nil {
			contentBody = body
			if value, ok := asString(front["description"]); ok && strings.TrimSpace(value) != "" {
				description = value
			}
		}
		content := fmt.Sprintf("---\ndescription: %s\n---\n%s", description, normalizeBody(string(contentBody)))
		name := strings.TrimSuffix(filepath.Base(path), ".md")
		if err := addImportOutput(out, filepath.ToSlash(filepath.Join(".agents", "agents", name+".md")), []byte(content)); err != nil {
			return nil, err
		}
	}

	toolsData, err := readMCPFile(root, filepath.Join(".mcp.json"))
	if err == nil {
		if err := addImportOutput(out, filepath.ToSlash(filepath.Join(".agents", "tools", "mcp.json")), toolsData); err != nil {
			return nil, err
		}
	} else if !errors.Is(err, fs.ErrNotExist) {
		return nil, err
	}

	settingsData, err := readClaudeSettings(root)
	if err == nil && len(settingsData) > 0 {
		if err := addImportOutput(out, filepath.ToSlash(filepath.Join(".agents", "hooks", "claude.json")), settingsData); err != nil {
			return nil, err
		}
	} else if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return nil, err
	}

	return out, nil
}

func parseCodexAgentTOML(data []byte) (codexAgentImport, error) {
	var info codexAgentImport
	scanner := bufio.NewScanner(strings.NewReader(string(data)))
	for scanner.Scan() {
		key, value, ok := parseTomlAssignment(scanner.Text())
		if !ok {
			continue
		}
		switch key {
		case "name":
			info.Name = value
		case "description":
			info.Description = value
		case "developer_instructions":
			info.Instructions = value
		}
	}
	if scanner.Err() != nil {
		return codexAgentImport{}, scanner.Err()
	}
	if info.Description == "" {
		info.Description = "Imported Codex agent"
	}
	if info.Instructions == "" {
		return codexAgentImport{}, errors.New("codex agent file missing developer_instructions")
	}
	return info, nil
}

var (
	codexServerRe         = regexp.MustCompile(`^\[mcp_servers\."([^"]+)"\]$`)
	codexServerBareRe     = regexp.MustCompile(`^\[mcp_servers\.([^\.\]]+)\]$`)
	codexServerEnvRe      = regexp.MustCompile(`^\[mcp_servers\."([^"]+)"\.env\]$`)
	codexServerEnvBareRe  = regexp.MustCompile(`^\[mcp_servers\.([^\.\]]+)\.env\]$`)
	codexServerHdrsRe     = regexp.MustCompile(`^\[mcp_servers\."([^"]+)"\.http_headers\]$`)
	codexServerHdrsBareRe = regexp.MustCompile(`^\[mcp_servers\.([^\.\]]+)\.http_headers\]$`)
)

func parseCodexConfigTOML(data []byte) (map[string]any, error) {
	type codexServer struct {
		Type    string
		Command string
		URL     string
		Args    []string
		Env     map[string]string
		Headers map[string]string
	}

	section := ""
	subsection := ""
	servers := map[string]*codexServer{}

	scanner := bufio.NewScanner(strings.NewReader(string(data)))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if match := codexServerRe.FindStringSubmatch(line); len(match) > 0 {
			section = match[1]
			subsection = ""
			if servers[section] == nil {
				servers[section] = &codexServer{}
			}
			continue
		}
		if match := codexServerBareRe.FindStringSubmatch(line); len(match) > 0 {
			section = match[1]
			subsection = ""
			if servers[section] == nil {
				servers[section] = &codexServer{}
			}
			continue
		}
		if match := codexServerEnvRe.FindStringSubmatch(line); len(match) > 0 {
			section = match[1]
			subsection = "env"
			if servers[section] == nil {
				servers[section] = &codexServer{}
			}
			if servers[section].Env == nil {
				servers[section].Env = map[string]string{}
			}
			continue
		}
		if match := codexServerEnvBareRe.FindStringSubmatch(line); len(match) > 0 {
			section = match[1]
			subsection = "env"
			if servers[section] == nil {
				servers[section] = &codexServer{}
			}
			if servers[section].Env == nil {
				servers[section].Env = map[string]string{}
			}
			continue
		}
		if match := codexServerHdrsRe.FindStringSubmatch(line); len(match) > 0 {
			section = match[1]
			subsection = "headers"
			if servers[section] == nil {
				servers[section] = &codexServer{}
			}
			if servers[section].Headers == nil {
				servers[section].Headers = map[string]string{}
			}
			continue
		}
		if match := codexServerHdrsBareRe.FindStringSubmatch(line); len(match) > 0 {
			section = match[1]
			subsection = "headers"
			if servers[section] == nil {
				servers[section] = &codexServer{}
			}
			if servers[section].Headers == nil {
				servers[section].Headers = map[string]string{}
			}
			continue
		}
		if section == "" {
			continue
		}
		key, rawValue, ok := parseTomlAssignment(line)
		if !ok {
			continue
		}
		server := servers[section]
		switch subsection {
		case "env":
			server.Env[key] = rawValue
		case "headers":
			server.Headers[key] = rawValue
		default:
			switch key {
			case "type":
				server.Type = rawValue
			case "command":
				server.Command = rawValue
			case "url":
				server.URL = rawValue
			case "args":
				values, err := parseTomlArray(rawValue)
				if err != nil {
					return nil, fmt.Errorf("invalid args for server %q: %w", section, err)
				}
				server.Args = values
			}
		}
	}
	if scanner.Err() != nil {
		return nil, scanner.Err()
	}

	output := map[string]any{"mcpServers": map[string]any{}}
	entries := output["mcpServers"].(map[string]any)
	for name, source := range servers {
		server := map[string]any{}
		if strings.TrimSpace(source.Type) != "" {
			server["type"] = source.Type
		}
		if strings.TrimSpace(source.Command) != "" {
			server["command"] = source.Command
		}
		if len(source.Args) > 0 {
			server["args"] = source.Args
		}
		if strings.TrimSpace(source.URL) != "" {
			server["url"] = source.URL
			if source.Command == "" && source.Type == "" {
				server["type"] = "streamable-http"
			}
		}
		if len(source.Env) > 0 {
			server["env"] = source.Env
		}
		if len(source.Headers) > 0 {
			server["headers"] = source.Headers
		}
		if source.Command == "" && source.URL == "" {
			return nil, fmt.Errorf("codex server %q has no command or url", name)
		}
		entries[name] = server
	}
	return output, nil
}

func readCodexConfig(root string) ([]byte, error) {
	data, err := os.ReadFile(filepath.Join(root, ".codex", "config.toml"))
	if errors.Is(err, fs.ErrNotExist) {
		return nil, fs.ErrNotExist
	}
	if err != nil {
		return nil, err
	}
	if _, err := parseCodexConfigTOML(data); err != nil {
		return nil, err
	}
	return data, nil
}

func parseTomlAssignment(line string) (key string, value string, ok bool) {
	eq := strings.Index(line, "=")
	if eq < 0 {
		return "", "", false
	}
	key = strings.TrimSpace(line[:eq])
	value = strings.TrimSpace(line[eq+1:])
	if key == "" || value == "" {
		return "", "", false
	}
	if strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"") {
		unquoted, err := strconv.Unquote(value)
		if err == nil {
			value = unquoted
		}
	}
	return key, value, true
}

func parseTomlArray(value string) ([]string, error) {
	trimmed := strings.TrimSpace(value)
	if !strings.HasPrefix(trimmed, "[") || !strings.HasSuffix(trimmed, "]") {
		return nil, errors.New("invalid array value")
	}
	items := strings.Split(trimmed[1:len(trimmed)-1], ",")
	out := make([]string, 0, len(items))
	for _, item := range items {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		unquoted, err := strconv.Unquote(item)
		if err != nil {
			return nil, err
		}
		out = append(out, unquoted)
	}
	return out, nil
}

func collectFilesWithSuffix(directory, suffix string) ([]string, error) {
	paths := make([]string, 0)
	info, err := os.Stat(directory)
	if errors.Is(err, fs.ErrNotExist) {
		return paths, nil
	}
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		return paths, nil
	}
	if err := filepath.WalkDir(directory, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			return nil
		}
		if strings.HasSuffix(entry.Name(), suffix) {
			paths = append(paths, path)
		}
		return nil
	}); err != nil {
		return nil, err
	}
	sort.Strings(paths)
	return paths, nil
}

func readMCPFile(root string, relativePath string) ([]byte, error) {
	data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(relativePath)))
	if err != nil {
		return nil, err
	}
	var parsed struct {
		Servers map[string]json.RawMessage `json:"mcpServers"`
	}
	if err := json.Unmarshal(data, &parsed); err != nil {
		return nil, err
	}
	if parsed.Servers == nil {
		return nil, errors.New("tools file missing mcpServers object")
	}
	return data, nil
}

func readClaudeSettings(root string) ([]byte, error) {
	data, err := os.ReadFile(filepath.Join(root, ".claude", "settings.json"))
	if err != nil {
		return nil, err
	}
	var parsed struct {
		Hooks map[string]json.RawMessage `json:"hooks"`
	}
	if err := json.Unmarshal(data, &parsed); err != nil {
		return nil, err
	}
	if parsed.Hooks == nil {
		return nil, nil
	}
	output, err := json.MarshalIndent(struct {
		Hooks map[string]json.RawMessage `json:"hooks"`
	}{Hooks: parsed.Hooks}, "", "  ")
	if err != nil {
		return nil, err
	}
	return output, nil
}

func normalizeMarkdownBody(data []byte) string {
	return normalizeBody(string(data))
}

func normalizeBody(body string) string {
	return strings.TrimRight(body, "\n") + "\n"
}

func stringSlice(value any) []string {
	if value == nil {
		return nil
	}
	raw, err := asStringSlice(value)
	if err == nil {
		paths := make([]string, 0, len(raw))
		for _, entry := range raw {
			if strings.TrimSpace(entry) != "" {
				paths = append(paths, entry)
			}
		}
		return paths
	}
	switch value := value.(type) {
	case string:
		pieces := splitGlobPatterns(value)
		if len(pieces) > 0 {
			return pieces
		}
		if strings.TrimSpace(value) != "" {
			return []string{strings.TrimSpace(value)}
		}
	}
	return nil
}

func asString(value any) (string, bool) {
	if cast, ok := value.(string); ok {
		return cast, true
	}
	return "", false
}

func asStringSlice(value any) ([]string, error) {
	raw, ok := value.([]any)
	if !ok {
		return nil, fmt.Errorf("expected array, got %T", value)
	}
	out := make([]string, 0, len(raw))
	for i, item := range raw {
		cast, ok := item.(string)
		if !ok {
			return nil, fmt.Errorf("expected string at index %d, got %T", i, item)
		}
		out = append(out, cast)
	}
	return out, nil
}
