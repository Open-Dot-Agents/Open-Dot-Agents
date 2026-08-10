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

	"github.com/pelletier/go-toml/v2"
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

	if err := importRootInstructions(root, "AGENTS.md", out); err != nil {
		return nil, err
	}
	if err := importCopilotInstructions(root, out); err != nil {
		return nil, err
	}
	if err := importCopilotProjectInstructions(root, out); err != nil {
		return nil, err
	}
	if err := importCopilotRules(root, out); err != nil {
		return nil, err
	}
	if err := importNestedCopilotRules(root, out); err != nil {
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
	if err := importCopilotSettings(root, out); err != nil {
		return nil, err
	}
	if err := importCopilotAllowedModels(root, out); err != nil {
		return nil, err
	}
	if err := importTree(root, filepath.Join(".github", "skills"), filepath.Join(".agents", "skills"), out, nil); err != nil {
		return nil, err
	}
	if err := importCopilotPrompts(root, out); err != nil {
		return nil, err
	}
	if err := importTree(root, filepath.Join(".github", "plugin"), filepath.Join(".agents", "plugins", "copilot"), out, nil); err != nil {
		return nil, err
	}
	return out, nil
}

func importRootInstructions(root, name string, out map[string][]byte) error {
	data, err := readSourceFile(filepath.Join(root, name))
	if errors.Is(err, fs.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	return addImportOutput(out, filepath.ToSlash(filepath.Join(".agents", "instructions", name)), data)
}

func importCopilotInstructions(root string, out map[string][]byte) error {
	data, err := readSourceFile(filepath.Join(root, ".github", "copilot-instructions.md"))
	if errors.Is(err, fs.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	return addImportOutput(out, filepath.ToSlash(filepath.Join(".agents", "instructions", "copilot-instructions.md")), data)
}

func importCopilotProjectInstructions(root string, out map[string][]byte) error {
	return importDiscoveredProjectFiles(root, filepath.Join(".agents", "instructions", "copilot-project"), out, func(relative string, entry fs.DirEntry) bool {
		if relative == "AGENTS.md" || relative == ".github/copilot-instructions.md" {
			return false
		}
		return entry.Name() == "AGENTS.md" || strings.HasSuffix(relative, "/.github/copilot-instructions.md")
	})
}

func importCopilotRules(root string, out map[string][]byte) error {
	sourceRoot := filepath.Join(root, ".github", "instructions")
	rules, err := collectFilesWithSuffix(sourceRoot, ".instructions.md")
	if err != nil {
		return err
	}
	for _, rule := range rules {
		data, err := os.ReadFile(rule)
		if err != nil {
			return err
		}
		relative, err := filepath.Rel(sourceRoot, rule)
		if err != nil {
			return err
		}
		name := strings.TrimSuffix(relative, ".instructions.md")
		content := data
		if _, _, err := frontMatter(data); err != nil {
			content = []byte(fmt.Sprintf("---\napplyTo: \"**/*\"\n---\n%s", normalizeMarkdownBody(data)))
		}
		if err := addImportOutput(out, filepath.ToSlash(filepath.Join(".agents", "rules", name+".md")), content); err != nil {
			return err
		}
	}
	return nil
}

func importNestedCopilotRules(root string, out map[string][]byte) error {
	return importDiscoveredProjectFiles(root, filepath.Join(".agents", "rules", "copilot-project"), out, func(relative string, entry fs.DirEntry) bool {
		return strings.HasSuffix(entry.Name(), ".instructions.md") && strings.Contains(relative, "/.github/instructions/")
	})
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
		content := data
		front, _, err := frontMatter(data)
		if err != nil {
			content = []byte(fmt.Sprintf("---\ndescription: %s\n---\n%s", name, normalizeMarkdownBody(data)))
		} else if description, ok := front["description"].(string); !ok || strings.TrimSpace(description) == "" {
			return fmt.Errorf("%s: front matter requires description", filepath.ToSlash(filepath.Clean(path)))
		}
		if err := addImportOutput(out, filepath.ToSlash(filepath.Join(".agents", "agents", name+".md")), content); err != nil {
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
	files := map[string][]byte{}
	for _, source := range []string{".mcp.json", ".github/mcp.json"} {
		data, err := readMCPFile(root, filepath.FromSlash(source))
		if errors.Is(err, fs.ErrNotExist) {
			continue
		}
		if err != nil {
			return err
		}
		var config any
		if err := json.Unmarshal(data, &config); err != nil {
			return err
		}
		if err := rejectInlineCredentials(config, source); err != nil {
			return err
		}
		files[source] = data
	}
	if len(files) == 0 {
		return nil
	}
	canonical := files[".github/mcp.json"]
	if len(canonical) == 0 {
		canonical = files[".mcp.json"]
	}
	if len(files) == 2 {
		merged := map[string]json.RawMessage{}
		for _, source := range []string{".mcp.json", ".github/mcp.json"} {
			var config struct {
				Servers map[string]json.RawMessage `json:"mcpServers"`
			}
			if err := json.Unmarshal(files[source], &config); err != nil {
				return err
			}
			for name, server := range config.Servers {
				merged[name] = server
			}
		}
		encoded, err := json.MarshalIndent(struct {
			Servers map[string]json.RawMessage `json:"mcpServers"`
		}{Servers: merged}, "", "  ")
		if err != nil {
			return err
		}
		canonical = append(encoded, '\n')
	}
	if err := addImportOutput(out, ".agents/tools/mcp.json", canonical); err != nil {
		return err
	}
	if _, hasRoot := files[".mcp.json"]; !hasRoot {
		return nil
	}
	rawFiles := make(map[string]string, len(files))
	for path, data := range files {
		rawFiles[path] = string(data)
	}
	provenance, err := json.MarshalIndent(copilotMCPProvenance{CanonicalSHA256: contentSHA256(canonical), Files: rawFiles}, "", "  ")
	if err != nil {
		return err
	}
	return addImportOutput(out, ".agents/settings/copilot-mcp-provenance.json", append(provenance, '\n'))
}

func importCopilotSettings(root string, out map[string][]byte) error {
	data, err := readSourceFile(filepath.Join(root, ".github", "copilot", "settings.json"))
	if errors.Is(err, fs.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	var settings any
	if err := json.Unmarshal(data, &settings); err != nil {
		return fmt.Errorf(".github/copilot/settings.json: %w", err)
	}
	if _, ok := settings.(map[string]any); !ok {
		return errors.New(".github/copilot/settings.json must contain a JSON object")
	}
	if err := rejectInlineCredentials(settings, ".github/copilot/settings.json"); err != nil {
		return err
	}
	return addImportOutput(out, ".agents/settings/copilot.json", data)
}

func importCopilotAllowedModels(root string, out map[string][]byte) error {
	data, err := readSourceFile(filepath.Join(root, ".github", "allowed_models.txt"))
	if errors.Is(err, fs.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	if len(strings.TrimSpace(string(data))) == 0 {
		return errors.New(".github/allowed_models.txt must not be empty")
	}
	return addImportOutput(out, ".agents/permissions/copilot-allowed-models.txt", data)
}

func importCopilotPrompts(root string, out map[string][]byte) error {
	paths, err := collectFilesWithSuffix(filepath.Join(root, ".github", "prompts"), ".prompt.md")
	if err != nil {
		return err
	}
	for _, path := range paths {
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		name := strings.TrimSuffix(filepath.Base(path), ".prompt.md")
		if err := addImportOutput(out, filepath.ToSlash(filepath.Join(".agents", "prompts", name+".md")), data); err != nil {
			return err
		}
	}
	return nil
}

func importTree(root, source, destination string, out map[string][]byte, accept func(string) bool) error {
	sourceRoot := filepath.Join(root, source)
	info, err := os.Lstat(sourceRoot)
	if errors.Is(err, fs.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("refusing symlinked import directory %s", sourceRoot)
	}
	if !info.IsDir() {
		return fmt.Errorf("%s is not a directory", filepath.ToSlash(source))
	}
	return filepath.WalkDir(sourceRoot, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			return nil
		}
		if entry.Type()&os.ModeSymlink != 0 {
			return fmt.Errorf("refusing symlinked import source %s", path)
		}
		relative, err := filepath.Rel(sourceRoot, path)
		if err != nil {
			return err
		}
		if accept != nil && !accept(relative) {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return addImportOutput(out, filepath.ToSlash(filepath.Join(destination, relative)), data)
	})
}

func importDiscoveredProjectFiles(root, destination string, out map[string][]byte, accept func(string, fs.DirEntry) bool) error {
	return filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		relative, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		relative = filepath.ToSlash(relative)
		if entry.IsDir() {
			if relative == ".git" || relative == ".agents" {
				return filepath.SkipDir
			}
			return nil
		}
		if !accept(relative, entry) {
			return nil
		}
		data, err := readSourceFile(path)
		if err != nil {
			return err
		}
		return addImportOutput(out, filepath.ToSlash(filepath.Join(destination, relative)), data)
	})
}

func importCodex(root string) (map[string][]byte, error) {
	out := make(map[string][]byte)
	configData, configErr := readCodexConfig(root)
	if configErr != nil && !errors.Is(configErr, fs.ErrNotExist) {
		return nil, configErr
	}

	if err := importRootInstructions(root, "AGENTS.md", out); err != nil {
		return nil, err
	}
	if err := importCodexProjectInstructions(root, configData, out); err != nil {
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
		var config map[string]any
		if err := toml.Unmarshal(data, &config); err != nil {
			return nil, fmt.Errorf("%s: %w", filepath.ToSlash(filepath.Clean(path)), err)
		}
		name := strings.TrimSuffix(filepath.Base(path), ".toml")
		canonicalName := name
		if configured, ok := config["name"].(string); ok && strings.TrimSpace(configured) != "" {
			canonicalName = configured
		}
		description, _ := config["description"].(string)
		if strings.TrimSpace(description) == "" {
			description = "Imported Codex agent"
		}
		instructions, _ := config["developer_instructions"].(string)
		if strings.TrimSpace(instructions) == "" {
			return nil, fmt.Errorf("%s: codex agent file missing developer_instructions", filepath.ToSlash(filepath.Clean(path)))
		}
		content := []byte(fmt.Sprintf("---\nname: %s\ndescription: %s\n---\n%s", strconv.Quote(canonicalName), strconv.Quote(description), normalizeBody(instructions)))
		if err := addImportOutput(out, filepath.ToSlash(filepath.Join(".agents", "agents", name+".md")), content); err != nil {
			return nil, err
		}
		if err := addImportOutput(out, filepath.ToSlash(filepath.Join(".agents", "settings", "codex-agents", filepath.Base(path))), data); err != nil {
			return nil, err
		}
	}
	if err := importTree(root, filepath.Join(".codex", "skills"), filepath.Join(".agents", "skills"), out, nil); err != nil {
		return nil, err
	}
	legacySkillFiles, err := collectRelativeImportFiles(filepath.Join(root, ".codex", "skills"))
	if err != nil {
		return nil, err
	}
	if len(legacySkillFiles) > 0 {
		encoded, err := json.MarshalIndent(codexLegacySkillsManifest{Files: legacySkillFiles}, "", "  ")
		if err != nil {
			return nil, err
		}
		if err := addImportOutput(out, ".agents/settings/codex-legacy-skills.json", append(encoded, '\n')); err != nil {
			return nil, err
		}
	}
	if err := importTree(root, filepath.Join(".codex", "rules"), filepath.Join(".agents", "permissions", "codex-rules"), out, func(path string) bool {
		return strings.HasSuffix(path, ".rules")
	}); err != nil {
		return nil, err
	}

	if !errors.Is(configErr, fs.ErrNotExist) {
		if err := importCodexConfig(configData, out); err != nil {
			return nil, err
		}
	}

	return out, nil
}

func collectRelativeImportFiles(root string) ([]string, error) {
	info, err := os.Lstat(root)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if info.Mode()&os.ModeSymlink != 0 || !info.IsDir() {
		return nil, fmt.Errorf("refusing invalid import directory %s", root)
	}
	var files []string
	err = filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil || entry.IsDir() {
			return walkErr
		}
		if entry.Type()&os.ModeSymlink != 0 {
			return fmt.Errorf("refusing symlinked import source %s", path)
		}
		relative, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		files = append(files, filepath.ToSlash(relative))
		return nil
	})
	sort.Strings(files)
	return files, err
}

func importCodexProjectInstructions(root string, configData []byte, out map[string][]byte) error {
	fallbacks := map[string]struct{}{}
	if len(configData) > 0 {
		var config map[string]any
		if err := toml.Unmarshal(configData, &config); err != nil {
			return err
		}
		for _, name := range stringSlice(config["project_doc_fallback_filenames"]) {
			if strings.TrimSpace(name) == "" || filepath.Base(name) != name {
				return fmt.Errorf(".codex/config.toml: project_doc_fallback_filenames entry %q must be a filename", name)
			}
			fallbacks[name] = struct{}{}
		}
	}
	return importDiscoveredProjectFiles(root, filepath.Join(".agents", "instructions", "codex-project"), out, func(relative string, entry fs.DirEntry) bool {
		if relative == "AGENTS.md" {
			return false
		}
		if entry.Name() == "AGENTS.md" || entry.Name() == "AGENTS.override.md" {
			return true
		}
		_, ok := fallbacks[entry.Name()]
		return ok
	})
}

func importClaude(root string) (map[string][]byte, error) {
	out := make(map[string][]byte)

	data, err := readSourceFile(filepath.Join(root, "CLAUDE.md"))
	if err == nil {
		if err := addImportOutput(out, filepath.ToSlash(filepath.Join(".agents", "instructions", "CLAUDE.md")), data); err != nil {
			return nil, err
		}
	} else if !errors.Is(err, fs.ErrNotExist) {
		return nil, err
	}

	claudeRulesRoot := filepath.Join(root, ".claude", "rules")
	rulePaths, err := collectFilesWithSuffix(claudeRulesRoot, ".md")
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
		relative, err := filepath.Rel(claudeRulesRoot, path)
		if err != nil {
			return nil, err
		}
		name := strings.TrimSuffix(relative, ".md")
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

func parseCodexConfigTOML(data []byte) (map[string]any, error) {
	var config map[string]any
	if err := toml.Unmarshal(data, &config); err != nil {
		return nil, err
	}
	servers, _ := stringMap(config["mcp_servers"])
	portable := map[string]any{}
	for name, value := range servers {
		source, ok := stringMap(value)
		if !ok {
			return nil, fmt.Errorf("codex server %q is not a table", name)
		}
		extension := make(map[string]any, len(source))
		for key, item := range source {
			extension[key] = item
		}
		for _, key := range []string{"command", "args", "env", "cwd", "url", "http_headers"} {
			delete(extension, key)
		}
		server := map[string]any{}
		if len(extension) > 0 {
			server["codex"] = extension
		}
		if command, ok := source["command"].(string); ok && command != "" {
			server["type"] = "stdio"
			server["command"] = command
			copyPortableField(source, server, "args", "env", "cwd")
		} else if url, ok := source["url"].(string); ok && url != "" {
			server["type"] = "streamable-http"
			server["url"] = url
			if headers, exists := source["http_headers"]; exists {
				server["headers"] = headers
			}
		} else {
			return nil, fmt.Errorf("codex server %q has no command or url", name)
		}
		portable[name] = server
	}
	return map[string]any{"mcpServers": portable}, nil
}

var codexConfigCategoryKeys = map[string][]string{
	"guardrails":  {"allow_login_shell", "sandbox_mode", "sandbox_workspace_write", "shell_environment_policy", "windows"},
	"hooks":       {"hooks"},
	"memories":    {"memories"},
	"permissions": {"approval_policy", "approvals_reviewer", "default_permissions", "include_permissions_instructions", "permissions"},
	"plugins":     {"marketplaces", "plugins"},
	"profiles":    {"profiles"},
}

func importCodexConfig(data []byte, out map[string][]byte) error {
	var config map[string]any
	if err := toml.Unmarshal(data, &config); err != nil {
		return err
	}
	if err := rejectInlineCredentials(config, ".codex/config.toml"); err != nil {
		return err
	}
	if err := addImportOutput(out, ".agents/settings/codex.raw.toml", data); err != nil {
		return err
	}
	remaining := make(map[string]any, len(config))
	for key, value := range config {
		remaining[key] = value
	}
	delete(remaining, "mcp_servers")
	for category, keys := range codexConfigCategoryKeys {
		fragment := map[string]any{}
		for _, key := range keys {
			if value, ok := remaining[key]; ok {
				fragment[key] = value
				delete(remaining, key)
			}
		}
		if err := addCodexFragment(out, category, fragment); err != nil {
			return err
		}
	}
	if err := addCodexFragment(out, "settings", remaining); err != nil {
		return err
	}
	if _, ok := config["mcp_servers"]; ok {
		portable, err := parseCodexConfigTOML(data)
		if err != nil {
			return err
		}
		encoded, err := json.MarshalIndent(portable, "", "  ")
		if err != nil {
			return err
		}
		if err := addImportOutput(out, ".agents/tools/mcp.json", append(encoded, '\n')); err != nil {
			return err
		}
	}
	return nil
}

func rejectInlineCredentials(value any, source string) error {
	return walkCredentialValues(value, source, "")
}

func walkCredentialValues(value any, source, path string) error {
	switch typed := value.(type) {
	case map[string]any:
		for key, item := range typed {
			keyPath := key
			if path != "" {
				keyPath = path + "." + key
			}
			isEnvironmentHeader := strings.HasSuffix(strings.ToLower(path), "env_http_headers")
			if text, ok := item.(string); ok && !isEnvironmentHeader && sensitiveCredentialKey(key) && !credentialReference(text) {
				return fmt.Errorf("%s contains an inline credential at %s; replace it with an environment-variable reference before import", source, keyPath)
			}
			if err := walkCredentialValues(item, source, keyPath); err != nil {
				return err
			}
		}
	case []any:
		for index, item := range typed {
			if err := walkCredentialValues(item, source, fmt.Sprintf("%s[%d]", path, index)); err != nil {
				return err
			}
		}
	}
	return nil
}

func sensitiveCredentialKey(key string) bool {
	normalized := strings.ToLower(strings.ReplaceAll(key, "-", "_"))
	if strings.HasSuffix(normalized, "_env_var") || strings.HasSuffix(normalized, "_env_vars") {
		return false
	}
	for _, marker := range []string{"authorization", "api_key", "password", "secret", "token"} {
		if strings.Contains(normalized, marker) {
			return true
		}
	}
	return false
}

func credentialReference(value string) bool {
	trimmed := strings.TrimSpace(value)
	return trimmed == "" || strings.Contains(trimmed, "$")
}

func addCodexFragment(out map[string][]byte, category string, fragment map[string]any) error {
	if len(fragment) == 0 {
		return nil
	}
	data, err := toml.Marshal(fragment)
	if err != nil {
		return err
	}
	return addImportOutput(out, filepath.ToSlash(filepath.Join(".agents", category, "codex.toml")), data)
}

func readCodexConfig(root string) ([]byte, error) {
	data, err := readSourceFile(filepath.Join(root, ".codex", "config.toml"))
	if errors.Is(err, fs.ErrNotExist) {
		return nil, fs.ErrNotExist
	}
	if err != nil {
		return nil, err
	}
	var config map[string]any
	if err := toml.Unmarshal(data, &config); err != nil {
		return nil, err
	}
	return data, nil
}

func collectFilesWithSuffix(directory, suffix string) ([]string, error) {
	paths := make([]string, 0)
	info, err := os.Lstat(directory)
	if errors.Is(err, fs.ErrNotExist) {
		return paths, nil
	}
	if err != nil {
		return nil, err
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return nil, fmt.Errorf("refusing symlinked import directory %s", directory)
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
			if entry.Type()&os.ModeSymlink != 0 {
				return fmt.Errorf("refusing symlinked import source %s", path)
			}
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
	data, err := readSourceFile(filepath.Join(root, filepath.FromSlash(relativePath)))
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
	data, err := readSourceFile(filepath.Join(root, ".claude", "settings.json"))
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
