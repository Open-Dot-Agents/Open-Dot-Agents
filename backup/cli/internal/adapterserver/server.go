package adapterserver

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/Open-Dot-Agents/Open-Dot-Agents/cli/internal/adapter"
	"github.com/Open-Dot-Agents/Open-Dot-Agents/cli/pkg/adapterprotocol"
	"gopkg.in/yaml.v3"
)

type Server struct {
	Target  string
	ID      string
	Name    string
	Version string
}

func New(target, version string) *Server {
	return &Server{
		Target:  target,
		ID:      "org.open-dot-agents." + target,
		Name:    "Open-Dot-Agents " + strings.ToUpper(target[:1]) + target[1:] + " reference adapter",
		Version: version,
	}
}

func (s *Server) Initialize(_ context.Context, params adapterprotocol.InitializeParams) (adapterprotocol.InitializeResult, error) {
	for _, offered := range params.ProtocolVersions {
		if offered == adapterprotocol.ProtocolVersion {
			return adapterprotocol.InitializeResult{ProtocolVersion: offered}, nil
		}
	}
	return adapterprotocol.InitializeResult{}, fmt.Errorf("no compatible adapter protocol")
}

func (s *Server) Describe(context.Context) (adapterprotocol.AdapterDescription, error) {
	var statuses map[string]string
	for _, info := range adapter.TargetInfos() {
		if info.ID == s.Target {
			statuses = info.CategoryStatuses
			break
		}
	}
	if statuses == nil {
		return adapterprotocol.AdapterDescription{}, fmt.Errorf("unknown reference target %q", s.Target)
	}
	return adapterprotocol.AdapterDescription{
		ID: s.ID, Name: s.Name, Version: s.Version,
		ProtocolVersion:  adapterprotocol.ProtocolVersion,
		Target:           s.Target,
		Capabilities:     []string{"validate", "import", "export"},
		CategoryStatuses: statuses,
		InputPatterns:    inputPatterns(s.Target),
		MaxSnapshotBytes: 16 << 20,
	}, nil
}

func (s *Server) Validate(ctx context.Context, params adapterprotocol.SnapshotParams) (adapterprotocol.OperationResult, error) {
	return s.operation(ctx, "export", params)
}

func (s *Server) ExportPlan(ctx context.Context, params adapterprotocol.SnapshotParams) (adapterprotocol.OperationResult, error) {
	return s.operation(ctx, "export", params)
}

func (s *Server) ImportPlan(ctx context.Context, params adapterprotocol.SnapshotParams) (adapterprotocol.OperationResult, error) {
	return s.operation(ctx, "import", params)
}

func (s *Server) operation(_ context.Context, operation string, params adapterprotocol.SnapshotParams) (adapterprotocol.OperationResult, error) {
	root, err := os.MkdirTemp("", "dota-adapter-")
	if err != nil {
		return adapterprotocol.OperationResult{}, err
	}
	defer os.RemoveAll(root)
	if err := materialize(root, params.Files); err != nil {
		return adapterprotocol.OperationResult{}, err
	}
	if err := hydrateReferenceExtensions(root, s.Target); err != nil {
		return adapterprotocol.OperationResult{}, err
	}
	if err := normalizeV1ForReferenceEngine(root); err != nil {
		return adapterprotocol.OperationResult{}, err
	}
	if err := normalizeMCPForReferenceEngine(root, s.Target); err != nil {
		return adapterprotocol.OperationResult{}, err
	}
	hookLosses, err := normalizeHooksForReferenceEngine(root, s.Target)
	if err != nil {
		return adapterprotocol.OperationResult{}, err
	}
	engine, err := adapter.NewForTarget(root, s.Target, true)
	if err != nil {
		return adapterprotocol.OperationResult{}, err
	}
	description, err := s.Describe(context.Background())
	if err != nil {
		return adapterprotocol.OperationResult{}, err
	}
	losses, err := unsupportedCategoryLosses(root, description.CategoryStatuses)
	if err != nil {
		return adapterprotocol.OperationResult{}, err
	}
	losses = append(losses, hookLosses...)
	if losses == nil {
		losses = []adapterprotocol.Loss{}
	}
	sort.Slice(losses, func(i, j int) bool {
		if losses[i].Path == losses[j].Path {
			return losses[i].Reason < losses[j].Reason
		}
		return losses[i].Path < losses[j].Path
	})
	var plan *adapter.GenerationPlan
	if operation == "import" {
		plan, err = engine.ImportPlan(true)
	} else {
		plan, err = engine.Plan(true)
	}
	if err != nil {
		return adapterprotocol.OperationResult{}, err
	}
	if operation == "import" {
		if err := normalizeImportPlan(plan, s.Target); err != nil {
			return adapterprotocol.OperationResult{}, err
		}
	}
	paths := make([]string, 0, len(plan.Outputs))
	for path := range plan.Outputs {
		paths = append(paths, filepath.ToSlash(path))
	}
	sort.Strings(paths)
	files := make([]adapterprotocol.File, 0, len(paths))
	for _, path := range paths {
		files = append(files, adapterprotocol.File{Path: path, Encoding: "base64", Content: base64.StdEncoding.EncodeToString(plan.Outputs[path])})
	}
	return adapterprotocol.OperationResult{
		Diagnostics: []adapterprotocol.Diagnostic{},
		Losses:      losses,
		Plan:        &adapterprotocol.Plan{Files: files},
	}, nil
}

func unsupportedCategoryLosses(root string, statuses map[string]string) ([]adapterprotocol.Loss, error) {
	var losses []adapterprotocol.Loss
	for category, status := range statuses {
		if status != "unsupported" {
			continue
		}
		categoryRoot := filepath.Join(root, ".agents", category)
		err := filepath.WalkDir(categoryRoot, func(path string, entry os.DirEntry, walkErr error) error {
			if walkErr != nil {
				if errors.Is(walkErr, os.ErrNotExist) {
					return nil
				}
				return walkErr
			}
			if entry.IsDir() || entry.Name() == "README.md" {
				return nil
			}
			relative, err := filepath.Rel(root, path)
			if err != nil {
				return err
			}
			losses = append(losses, adapterprotocol.Loss{
				Path: filepath.ToSlash(relative), Reason: "target does not support the category", Severity: "warning",
			})
			return nil
		})
		if err != nil {
			return nil, err
		}
	}
	return losses, nil
}

func normalizeHooksForReferenceEngine(root, target string) ([]adapterprotocol.Loss, error) {
	hooksRoot := filepath.Join(root, ".agents", "hooks")
	entries, err := os.ReadDir(hooksRoot)
	if errors.Is(err, os.ErrNotExist) {
		return []adapterprotocol.Loss{}, nil
	}
	if err != nil {
		return nil, err
	}
	var losses []adapterprotocol.Loss
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}
		path := filepath.Join(hooksRoot, entry.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		var probe struct {
			Hooks json.RawMessage `json:"hooks"`
		}
		if err := json.Unmarshal(data, &probe); err != nil {
			return nil, err
		}
		if len(probe.Hooks) == 0 || probe.Hooks[0] != '[' {
			continue
		}
		if target != "claude" {
			relative := filepath.ToSlash(filepath.Join(".agents", "hooks", entry.Name()))
			losses = append(losses, adapterprotocol.Loss{Path: relative, Reason: "portable hooks are not representable by this reference adapter", Severity: "warning"})
			if err := os.Remove(path); err != nil {
				return nil, err
			}
			continue
		}
		converted, conversionLosses, err := convertPortableHooksToClaude(entry.Name(), probe.Hooks)
		if err != nil {
			return nil, err
		}
		losses = append(losses, conversionLosses...)
		if err := os.WriteFile(path, converted, 0o644); err != nil {
			return nil, err
		}
	}
	return losses, nil
}

func convertPortableHooksToClaude(name string, raw json.RawMessage) ([]byte, []adapterprotocol.Loss, error) {
	var hooks []struct {
		ID      string         `json:"id"`
		Event   string         `json:"event"`
		Matcher map[string]any `json:"matcher"`
		Action  struct {
			Type    string   `json:"type"`
			Command []string `json:"command"`
		} `json:"action"`
	}
	if err := json.Unmarshal(raw, &hooks); err != nil {
		return nil, nil, err
	}
	events := map[string][]any{}
	var losses []adapterprotocol.Loss
	eventNames := map[string]string{
		"session-start": "SessionStart", "session-end": "SessionEnd", "before-tool": "PreToolUse",
		"after-tool": "PostToolUse", "before-compaction": "PreCompact",
	}
	for _, hook := range hooks {
		path := filepath.ToSlash(filepath.Join(".agents", "hooks", name)) + "#" + hook.ID
		event := eventNames[hook.Event]
		if event == "" || hook.Action.Type != "command" || len(hook.Action.Command) == 0 {
			losses = append(losses, adapterprotocol.Loss{Path: path, Reason: "hook event or action is not representable in Claude", Severity: "warning"})
			continue
		}
		definition := map[string]any{"hooks": []any{map[string]any{"type": "command", "command": shellJoin(hook.Action.Command)}}}
		if matcher, ok := hook.Matcher["tool"].(string); ok && matcher != "" {
			definition["matcher"] = matcher
		} else if len(hook.Matcher) > 0 {
			losses = append(losses, adapterprotocol.Loss{Path: path + "/matcher", Reason: "matcher is not representable in Claude", Severity: "warning"})
		}
		events[event] = append(events[event], definition)
	}
	encoded, err := json.MarshalIndent(struct {
		Hooks map[string][]any `json:"hooks"`
	}{Hooks: events}, "", "  ")
	if err != nil {
		return nil, nil, err
	}
	return append(encoded, '\n'), losses, nil
}

func shellJoin(arguments []string) string {
	quoted := make([]string, len(arguments))
	for index, argument := range arguments {
		quoted[index] = "'" + strings.ReplaceAll(argument, "'", "'\"'\"'") + "'"
	}
	return strings.Join(quoted, " ")
}

func normalizeImportPlan(plan *adapter.GenerationPlan, target string) error {
	if plan == nil {
		return errors.New("reference engine returned no import plan")
	}
	normalized := make(map[string][]byte, len(plan.Outputs))
	for output, data := range plan.Outputs {
		path := filepath.ToSlash(output)
		parts := strings.Split(path, "/")
		if len(parts) >= 3 && parts[0] == ".agents" && isVendorStructuredCategory(parts[1]) {
			path = strings.Join(append([]string{".agents", "extensions", "org.open-dot-agents." + target}, parts[1:]...), "/")
		}
		var err error
		if strings.HasPrefix(path, ".agents/rules/") && strings.HasSuffix(path, ".md") {
			data, err = normalizeImportedRule(data)
		}
		if path == ".agents/tools/mcp.json" {
			data, err = normalizeImportedMCP(data)
		}
		if err != nil {
			return fmt.Errorf("normalize import output %s: %w", path, err)
		}
		if _, exists := normalized[path]; exists {
			return fmt.Errorf("normalized import output collision %q", path)
		}
		normalized[path] = data
	}
	plan.Outputs = normalized
	return nil
}

func isVendorStructuredCategory(category string) bool {
	switch category {
	case "guardrails", "hooks", "memories", "permissions", "plugins", "profiles", "settings":
		return true
	default:
		return false
	}
}

func normalizeImportedRule(data []byte) ([]byte, error) {
	text := strings.ReplaceAll(string(data), "\r\n", "\n")
	if !strings.HasPrefix(text, "---\n") {
		return data, nil
	}
	end := strings.Index(text[4:], "\n---\n")
	if end < 0 {
		return nil, errors.New("unterminated rule front matter")
	}
	front := map[string]any{}
	if err := yaml.Unmarshal([]byte(text[4:4+end]), &front); err != nil {
		return nil, err
	}
	if scalar, ok := front["applyTo"].(string); ok {
		items := strings.Split(scalar, ",")
		patterns := make([]string, 0, len(items))
		for _, item := range items {
			if trimmed := strings.TrimSpace(item); trimmed != "" {
				patterns = append(patterns, trimmed)
			}
		}
		front["applyTo"] = patterns
	}
	encoded, err := yaml.Marshal(front)
	if err != nil {
		return nil, err
	}
	normalized := append([]byte("---\n"), encoded...)
	normalized = append(normalized, []byte("---\n")...)
	normalized = append(normalized, []byte(text[4+end+5:])...)
	return normalized, nil
}

func normalizeImportedMCP(data []byte) ([]byte, error) {
	var document struct {
		Servers map[string]map[string]any `json:"mcpServers"`
	}
	if err := json.Unmarshal(data, &document); err != nil {
		return nil, err
	}
	changed := false
	for _, server := range document.Servers {
		if server["type"] == "http" {
			server["type"] = "streamable-http"
			changed = true
		}
	}
	if !changed {
		return data, nil
	}
	encoded, err := json.MarshalIndent(document, "", "  ")
	if err != nil {
		return nil, err
	}
	return append(encoded, '\n'), nil
}

func hydrateReferenceExtensions(root, target string) error {
	source := filepath.Join(root, ".agents", "extensions", "org.open-dot-agents."+target)
	entries, err := os.ReadDir(source)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	for _, category := range entries {
		if !category.IsDir() {
			continue
		}
		categoryRoot := filepath.Join(source, category.Name())
		err := filepath.WalkDir(categoryRoot, func(path string, entry os.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			if entry.IsDir() {
				return nil
			}
			if entry.Type()&os.ModeSymlink != 0 {
				return fmt.Errorf("refusing symlinked extension file %s", path)
			}
			relative, err := filepath.Rel(categoryRoot, path)
			if err != nil {
				return err
			}
			destination := filepath.Join(root, ".agents", category.Name(), relative)
			data, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			if err := os.MkdirAll(filepath.Dir(destination), 0o755); err != nil {
				return err
			}
			return os.WriteFile(destination, data, 0o644)
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func normalizeMCPForReferenceEngine(root, target string) error {
	if target == "codex" {
		return nil
	}
	path := filepath.Join(root, ".agents", "tools", "mcp.json")
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	var document struct {
		Servers map[string]map[string]any `json:"mcpServers"`
	}
	if err := json.Unmarshal(data, &document); err != nil {
		return err
	}
	changed := false
	for _, server := range document.Servers {
		if server["type"] == "streamable-http" {
			server["type"] = "http"
			changed = true
		}
	}
	if !changed {
		return nil
	}
	encoded, err := json.MarshalIndent(document, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(encoded, '\n'), 0o644)
}

func normalizeV1ForReferenceEngine(root string) error {
	rulesRoot := filepath.Join(root, ".agents", "rules")
	return filepath.WalkDir(rulesRoot, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			if errors.Is(walkErr, os.ErrNotExist) {
				return nil
			}
			return walkErr
		}
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".md" || entry.Name() == "README.md" {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		text := strings.ReplaceAll(string(data), "\r\n", "\n")
		if !strings.HasPrefix(text, "---\n") {
			return nil
		}
		end := strings.Index(text[4:], "\n---\n")
		if end < 0 {
			return nil
		}
		front := map[string]any{}
		if err := yaml.Unmarshal([]byte(text[4:4+end]), &front); err != nil {
			return err
		}
		items, ok := front["applyTo"].([]any)
		if !ok {
			return nil
		}
		patterns := make([]string, 0, len(items))
		for _, item := range items {
			value, ok := item.(string)
			if !ok {
				return fmt.Errorf("rule %s applyTo must contain strings", path)
			}
			patterns = append(patterns, value)
		}
		front["applyTo"] = strings.Join(patterns, ", ")
		encoded, err := yaml.Marshal(front)
		if err != nil {
			return err
		}
		normalized := append([]byte("---\n"), encoded...)
		normalized = append(normalized, []byte("---\n")...)
		normalized = append(normalized, []byte(text[4+end+5:])...)
		return os.WriteFile(path, normalized, 0o644)
	})
}

func materialize(root string, files []adapterprotocol.File) error {
	var total int64
	for _, file := range files {
		if err := validatePath(file.Path); err != nil {
			return err
		}
		var data []byte
		var err error
		switch file.Encoding {
		case "utf-8":
			data = []byte(file.Content)
		case "base64":
			data, err = base64.StdEncoding.DecodeString(file.Content)
		default:
			err = fmt.Errorf("unsupported snapshot encoding %q", file.Encoding)
		}
		if err != nil {
			return err
		}
		total += int64(len(data))
		if total > 16<<20 {
			return errors.New("snapshot exceeds adapter limit")
		}
		destination := filepath.Join(root, filepath.FromSlash(file.Path))
		if err := os.MkdirAll(filepath.Dir(destination), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(destination, data, 0o644); err != nil {
			return err
		}
	}
	return nil
}

func validatePath(path string) error {
	clean := filepath.ToSlash(filepath.Clean(filepath.FromSlash(path)))
	if path == "" || filepath.IsAbs(path) || clean != path || clean == "." || clean == ".." || strings.HasPrefix(clean, "../") {
		return fmt.Errorf("unsafe snapshot path %q", path)
	}
	return nil
}

func inputPatterns(target string) []string {
	base := []string{".agents/**"}
	switch target {
	case "copilot":
		return append(base, "AGENTS.md", "**/AGENTS.md", ".github/**", ".mcp.json")
	case "codex":
		return append(base, "AGENTS.md", "**/AGENTS.md", "**/AGENTS.override.md", ".codex/**")
	case "claude":
		return append(base, "CLAUDE.md", "**/CLAUDE.md", ".claude/**", ".mcp.json")
	default:
		return base
	}
}
