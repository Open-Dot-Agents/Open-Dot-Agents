package dotacore

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/Open-Dot-Agents/Open-Dot-Agents/cli/internal/specdata"
	jsonschema "github.com/santhosh-tekuri/jsonschema/v5"
	"gopkg.in/yaml.v3"
)

const ManifestPath = ".agents/manifest.json"

var (
	reverseDNSPattern = regexp.MustCompile(`^[a-z0-9][a-z0-9-]*(\.[a-z0-9][a-z0-9-]*)+$`)
	skillNamePattern  = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)
	categoryNames     = map[string]bool{
		"agents": true, "guardrails": true, "hooks": true, "instructions": true,
		"memories": true, "permissions": true, "plugins": true, "profiles": true,
		"prompts": true, "rules": true, "settings": true, "skills": true, "tools": true,
	}
)

type Manifest struct {
	Schema      string         `json:"$schema,omitempty"`
	SpecVersion string         `json:"specVersion"`
	Conformance []string       `json:"conformance"`
	Extensions  map[string]any `json:"extensions,omitempty"`
}

type ValidationResult struct {
	Manifest    Manifest
	Diagnostics []Diagnostic
}

type Diagnostic struct {
	Code     string `json:"code"`
	Severity string `json:"severity"`
	Message  string `json:"message"`
	Path     string `json:"path,omitempty"`
}

func Validate(root string) (ValidationResult, error) {
	data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(ManifestPath)))
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return ValidationResult{}, diagnosticError("DOTA1000", ManifestPath, ".agents/manifest.json is required")
		}
		return ValidationResult{}, err
	}
	if bytes.Contains(data, []byte(`"format_version"`)) {
		return ValidationResult{}, diagnosticError("DOTA1001", ManifestPath, "v0.0.1 is unsupported; replace the tree with the v1 contract")
	}
	if err := validateEmbeddedSchema("manifest.schema.json", data); err != nil {
		return ValidationResult{}, diagnosticError("DOTA1002", ManifestPath, err.Error())
	}
	var manifest Manifest
	if err := decodeStrictJSON(data, &manifest); err != nil {
		return ValidationResult{}, diagnosticError("DOTA1002", ManifestPath, err.Error())
	}
	if err := validateTree(root); err != nil {
		return ValidationResult{}, err
	}
	if lockData, lockErr := os.ReadFile(filepath.Join(root, filepath.FromSlash(LockPath))); lockErr == nil {
		if err := validateEmbeddedSchema("adapters-lock.schema.json", lockData); err != nil {
			return ValidationResult{}, diagnosticError("DOTA4000", LockPath, err.Error())
		}
	} else if !errors.Is(lockErr, fs.ErrNotExist) {
		return ValidationResult{}, lockErr
	}
	return ValidationResult{Manifest: manifest}, nil
}

func validateEmbeddedSchema(name string, data []byte) error {
	compiler := jsonschema.NewCompiler()
	for _, resource := range []string{"common.schema.json", name} {
		contents, err := specdata.Schemas.ReadFile("schema/" + resource)
		if err != nil {
			return err
		}
		url := "https://open-dot-agents.github.io/Open-Dot-Agents/spec/v1/schema/" + resource
		if err := compiler.AddResource(url, bytes.NewReader(contents)); err != nil {
			return err
		}
	}
	schema, err := compiler.Compile("https://open-dot-agents.github.io/Open-Dot-Agents/spec/v1/schema/" + name)
	if err != nil {
		return err
	}
	var document any
	if err := json.Unmarshal(data, &document); err != nil {
		return err
	}
	return schema.Validate(document)
}

func validateTree(root string) error {
	base := filepath.Join(root, ".agents")
	entries, err := os.ReadDir(base)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		name := entry.Name()
		if name == "manifest.json" || name == "adapters.lock.json" || name == ".dota" {
			continue
		}
		if name == "extensions" {
			if err := validateExtensions(filepath.Join(base, name)); err != nil {
				return err
			}
			continue
		}
		if !categoryNames[name] || !entry.IsDir() {
			if entry.IsDir() {
				empty, emptyErr := directoryEmpty(filepath.Join(base, name))
				if emptyErr != nil {
					return emptyErr
				}
				if empty {
					continue
				}
			}
			return diagnosticError("DOTA1100", filepath.ToSlash(filepath.Join(".agents", name)), "unknown v1 root entry")
		}
	}
	checks := []struct {
		category string
		fn       func(string) error
	}{
		{"instructions", validateInstructions}, {"rules", validateRules}, {"agents", validateAgents},
		{"prompts", validatePrompts}, {"skills", validateSkills},
		{"memories", structuredValidator("memory.schema.json")},
		{"hooks", structuredValidator("hooks.schema.json")},
		{"permissions", structuredValidator("policy.schema.json")},
		{"guardrails", structuredValidator("guardrails.schema.json")},
		{"profiles", validateProfiles},
		{"settings", structuredValidator("settings.schema.json")},
		{"plugins", structuredValidator("plugin.schema.json")},
	}
	for _, check := range checks {
		path := filepath.Join(base, check.category)
		if err := check.fn(path); err != nil && !errors.Is(err, fs.ErrNotExist) {
			return err
		}
	}
	if err := validateMCP(filepath.Join(base, "tools", "mcp.json")); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return err
	}
	return nil
}

func directoryEmpty(root string) (bool, error) {
	empty := true
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if path != root && !entry.IsDir() {
			empty = false
			return fs.SkipAll
		}
		return nil
	})
	return empty, err
}

func validateExtensions(root string) error {
	entries, err := os.ReadDir(root)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if !entry.IsDir() || !reverseDNSPattern.MatchString(entry.Name()) {
			return diagnosticError("DOTA1101", filepath.ToSlash(filepath.Join(".agents/extensions", entry.Name())), "extension directory must use a reverse-DNS identifier")
		}
		namespace := filepath.Join(root, entry.Name())
		if err := filepath.WalkDir(namespace, func(path string, child fs.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			if child.Type()&os.ModeSymlink != 0 {
				return diagnosticError("DOTA1101", filepath.ToSlash(path), "symlinked extension entries are forbidden")
			}
			return nil
		}); err != nil {
			return err
		}
	}
	return nil
}

func validateInstructions(root string) error {
	return walkMarkdown(root, func(path string, front map[string]any) error {
		if priority, ok := front["priority"]; ok {
			switch value := priority.(type) {
			case int, int64, uint64, float64:
				if number, ok := value.(float64); ok && number != math.Trunc(number) {
					return diagnosticError("DOTA1200", path, "instruction priority must be an integer")
				}
			default:
				return diagnosticError("DOTA1200", path, "instruction priority must be an integer")
			}
		}
		return nil
	})
}

func validateRules(root string) error {
	seen := map[string]string{}
	return walkMarkdown(root, func(path string, front map[string]any) error {
		id := strings.TrimSpace(stringValue(front["id"]))
		patterns := stringSlice(front["applyTo"])
		if id == "" || len(patterns) == 0 {
			return diagnosticError("DOTA1201", path, "rule front matter requires id and applyTo array")
		}
		if previous, exists := seen[id]; exists {
			return diagnosticError("DOTA1201", path, fmt.Sprintf("rule id %q is also defined in %s", id, previous))
		}
		seen[id] = path
		if raw, exists := front["exclude"]; exists && len(stringSlice(raw)) == 0 {
			return diagnosticError("DOTA1201", path, "rule exclude must be an array of patterns")
		}
		for _, pattern := range append(patterns, stringSlice(front["exclude"])...) {
			clean := filepath.ToSlash(filepath.Clean(filepath.FromSlash(pattern)))
			if strings.HasPrefix(pattern, "/") || clean == ".." || strings.HasPrefix(clean, "../") || strings.Contains(pattern, "\\") {
				return diagnosticError("DOTA1201", path, "rule patterns must be repository-relative")
			}
		}
		return nil
	})
}

func validateAgents(root string) error {
	return walkMarkdown(root, func(path string, front map[string]any) error {
		if strings.TrimSpace(stringValue(front["name"])) == "" || strings.TrimSpace(stringValue(front["description"])) == "" {
			return diagnosticError("DOTA1202", path, "agent front matter requires name and description")
		}
		return nil
	})
}

func validatePrompts(root string) error {
	return walkMarkdown(root, func(path string, front map[string]any) error {
		if strings.TrimSpace(stringValue(front["name"])) == "" || strings.TrimSpace(stringValue(front["description"])) == "" {
			return diagnosticError("DOTA1203", path, "prompt front matter requires name and description")
		}
		return nil
	})
}

func validateSkills(root string) error {
	entries, err := os.ReadDir(root)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if entry.Name() == "README.md" {
			continue
		}
		if !entry.IsDir() || !skillNamePattern.MatchString(entry.Name()) || len(entry.Name()) > 64 {
			return diagnosticError("DOTA1204", filepath.ToSlash(filepath.Join(".agents/skills", entry.Name())), "invalid Agent Skills directory name")
		}
		path := filepath.Join(root, entry.Name(), "SKILL.md")
		data, err := readRegular(path)
		if err != nil {
			return diagnosticError("DOTA1205", filepath.ToSlash(path), "skill requires SKILL.md")
		}
		front, err := parseFrontMatter(data, true)
		if err != nil {
			return diagnosticError("DOTA1205", filepath.ToSlash(path), err.Error())
		}
		name := stringValue(front["name"])
		description := stringValue(front["description"])
		if name != entry.Name() || len(description) == 0 || len(description) > 1024 {
			return diagnosticError("DOTA1205", filepath.ToSlash(path), "skill name must match its directory and description must contain 1-1024 characters")
		}
		if err := filepath.WalkDir(filepath.Join(root, entry.Name()), func(path string, child fs.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			if child.Type()&os.ModeSymlink != 0 {
				return diagnosticError("DOTA1206", filepath.ToSlash(path), "symlinked skill entries are forbidden")
			}
			return nil
		}); err != nil {
			return err
		}
	}
	return nil
}

func structuredValidator(schemaName string) func(string) error {
	return func(root string) error { return validateJSONFiles(root, schemaName, nil) }
}

func validateJSONFiles(root, schemaName string, visit func(string, []byte) error) error {
	return filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			return nil
		}
		if entry.Type()&os.ModeSymlink != 0 {
			return diagnosticError("DOTA1300", filepath.ToSlash(path), "symlinked category files are forbidden")
		}
		if entry.Name() == "README.md" {
			return nil
		}
		if filepath.Ext(entry.Name()) != ".json" {
			return diagnosticError("DOTA1301", filepath.ToSlash(path), "v1 structured category files must use .json")
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		var value any
		if err := decodeStrictJSON(data, &value); err != nil {
			return diagnosticError("DOTA1302", filepath.ToSlash(path), err.Error())
		}
		if err := validateEmbeddedSchema(schemaName, data); err != nil {
			return diagnosticError("DOTA1302", filepath.ToSlash(path), err.Error())
		}
		if visit != nil {
			return visit(filepath.ToSlash(path), data)
		}
		return nil
	})
}

func validateProfiles(root string) error {
	parents := map[string]string{}
	profileFiles := map[string]string{}
	err := validateJSONFiles(root, "profiles.schema.json", func(path string, data []byte) error {
		var document struct {
			Profiles map[string]struct {
				Extends string         `json:"extends"`
				Patch   map[string]any `json:"patch"`
			} `json:"profiles"`
		}
		if err := decodeStrictJSON(data, &document); err != nil {
			return err
		}
		for name, profile := range document.Profiles {
			if previous, exists := profileFiles[name]; exists {
				return diagnosticError("DOTA1305", path, fmt.Sprintf("profile %q is also defined in %s", name, previous))
			}
			profileFiles[name] = path
			parents[name] = profile.Extends
		}
		return nil
	})
	if err != nil {
		return err
	}
	for name, parent := range parents {
		if parent != "" {
			if _, exists := parents[parent]; !exists {
				return diagnosticError("DOTA1305", profileFiles[name], fmt.Sprintf("profile %q extends unknown profile %q", name, parent))
			}
		}
	}
	state := map[string]uint8{}
	var visit func(string) error
	visit = func(name string) error {
		switch state[name] {
		case 1:
			return diagnosticError("DOTA1305", profileFiles[name], "profile inheritance cycle includes "+name)
		case 2:
			return nil
		}
		state[name] = 1
		if parent := parents[name]; parent != "" {
			if err := visit(parent); err != nil {
				return err
			}
		}
		state[name] = 2
		return nil
	}
	for name := range parents {
		if err := visit(name); err != nil {
			return err
		}
	}
	return nil
}

func validateMCP(path string) error {
	data, err := readRegular(path)
	if err != nil {
		return err
	}
	var document struct {
		Servers map[string]map[string]any `json:"mcpServers"`
	}
	if err := decodeStrictJSON(data, &document); err != nil {
		return diagnosticError("DOTA1303", filepath.ToSlash(path), err.Error())
	}
	if err := validateEmbeddedSchema("mcp.schema.json", data); err != nil {
		return diagnosticError("DOTA1303", filepath.ToSlash(path), err.Error())
	}
	if document.Servers == nil {
		return diagnosticError("DOTA1303", filepath.ToSlash(path), "mcpServers object is required")
	}
	for name, server := range document.Servers {
		typeName := stringValue(server["type"])
		if typeName == "" {
			if _, ok := server["command"]; ok {
				typeName = "stdio"
			}
		}
		if typeName != "stdio" && typeName != "streamable-http" {
			return diagnosticError("DOTA1303", filepath.ToSlash(path), fmt.Sprintf("MCP server %q uses unsupported transport %q", name, typeName))
		}
		for _, field := range []string{"env", "headers"} {
			if values, ok := server[field].(map[string]any); ok {
				for key, raw := range values {
					value := stringValue(raw)
					sensitive := regexp.MustCompile(`(?i)(token|key|secret|password|credential)`).MatchString(key)
					if sensitive && !regexp.MustCompile(`^\$\{[A-Z_][A-Z0-9_]*\}$`).MatchString(value) {
						return diagnosticError("DOTA1304", filepath.ToSlash(path), fmt.Sprintf("%s.%s must use an environment reference", field, key))
					}
				}
			}
		}
	}
	return nil
}

func walkMarkdown(root string, validate func(string, map[string]any) error) error {
	return filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			return nil
		}
		if entry.Type()&os.ModeSymlink != 0 {
			return diagnosticError("DOTA1206", filepath.ToSlash(path), "symlinked Markdown is forbidden")
		}
		if entry.Name() == "README.md" {
			return nil
		}
		if filepath.Ext(entry.Name()) != ".md" {
			return diagnosticError("DOTA1207", filepath.ToSlash(path), "narrative category files must use .md")
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		front, err := parseFrontMatter(data, false)
		if err != nil {
			return diagnosticError("DOTA1208", filepath.ToSlash(path), err.Error())
		}
		return validate(filepath.ToSlash(path), front)
	})
}

func parseFrontMatter(data []byte, required bool) (map[string]any, error) {
	text := strings.ReplaceAll(string(data), "\r\n", "\n")
	if !strings.HasPrefix(text, "---\n") {
		if required {
			return nil, errors.New("YAML front matter is required")
		}
		return map[string]any{}, nil
	}
	end := strings.Index(text[4:], "\n---\n")
	if end < 0 {
		return nil, errors.New("unterminated YAML front matter")
	}
	front := map[string]any{}
	if err := yaml.Unmarshal([]byte(text[4:4+end]), &front); err != nil {
		return nil, err
	}
	return front, nil
}

func readRegular(path string) ([]byte, error) {
	info, err := os.Lstat(path)
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() {
		return nil, errors.New("not a regular file")
	}
	return os.ReadFile(path)
}

func decodeStrictJSON(data []byte, target any) error {
	if err := rejectDuplicateKeys(data); err != nil {
		return err
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return err
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return errors.New("multiple JSON values are forbidden")
	}
	return nil
}

func rejectDuplicateKeys(data []byte) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	var parseValue func() error
	parseValue = func() error {
		token, err := decoder.Token()
		if err != nil {
			return err
		}
		delim, ok := token.(json.Delim)
		if !ok {
			return nil
		}
		switch delim {
		case '{':
			seen := map[string]bool{}
			for decoder.More() {
				keyToken, err := decoder.Token()
				if err != nil {
					return err
				}
				key, ok := keyToken.(string)
				if !ok {
					return errors.New("invalid JSON object key")
				}
				if seen[key] {
					return fmt.Errorf("duplicate JSON key %q", key)
				}
				seen[key] = true
				if err := parseValue(); err != nil {
					return err
				}
			}
			_, err = decoder.Token()
			return err
		case '[':
			for decoder.More() {
				if err := parseValue(); err != nil {
					return err
				}
			}
			_, err = decoder.Token()
			return err
		default:
			return errors.New("invalid JSON delimiter")
		}
	}
	if err := parseValue(); err != nil {
		return err
	}
	if _, err := decoder.Token(); !errors.Is(err, io.EOF) {
		return errors.New("multiple JSON values are forbidden")
	}
	return nil
}

func stringValue(value any) string {
	result, _ := value.(string)
	return result
}

func stringSlice(value any) []string {
	raw, ok := value.([]any)
	if !ok {
		return nil
	}
	result := make([]string, 0, len(raw))
	for _, item := range raw {
		value, ok := item.(string)
		if !ok || strings.TrimSpace(value) == "" {
			return nil
		}
		result = append(result, value)
	}
	return result
}

func diagnosticError(code, path, message string) error {
	return fmt.Errorf("%s %s: %s", code, path, message)
}

func CategoryList(root string) ([]string, error) {
	entries, err := os.ReadDir(filepath.Join(root, ".agents"))
	if err != nil {
		return nil, err
	}
	var categories []string
	for _, entry := range entries {
		if entry.IsDir() && categoryNames[entry.Name()] {
			categories = append(categories, entry.Name())
		}
	}
	sort.Strings(categories)
	return categories, nil
}
