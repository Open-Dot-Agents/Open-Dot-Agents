package adapter

import (
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

type copilotRenderer struct{}

func (copilotRenderer) target() string {
	return "copilot"
}

func (copilotRenderer) manifestDirectory() string {
	return ".github"
}

func (copilotRenderer) unsupportedCategories() []string {
	return []string{"guardrails", "memories", "profiles"}
}

func (copilotRenderer) render(base string) (map[string][]byte, error) {
	out := make(map[string][]byte)
	if err := copilotInstructions(base, out); err != nil {
		return nil, err
	}
	if err := copyDirectory(filepath.Join(base, "instructions", "copilot-project"), ".", out); err != nil {
		return nil, err
	}
	if err := rules(base, ".github/instructions", ".instructions.md", out); err != nil {
		return nil, err
	}
	if err := copyDirectory(filepath.Join(base, "rules", "copilot-project"), ".", out); err != nil {
		return nil, err
	}
	if err := agents(base, ".github/agents", ".agent.md", out); err != nil {
		return nil, err
	}
	if err := copyJSON(base, "hooks", ".github/hooks", out); err != nil {
		return nil, err
	}
	if err := copilotMCP(base, out); err != nil {
		return nil, err
	}
	if err := copilotSettings(base, out); err != nil {
		return nil, err
	}
	if err := copilotAllowedModels(base, out); err != nil {
		return nil, err
	}
	if err := copySkills(base, ".github/skills", out); err != nil {
		return nil, err
	}
	if err := copilotPrompts(base, out); err != nil {
		return nil, err
	}
	if err := copyDirectory(filepath.Join(base, "plugins", "copilot"), ".github/plugin", out); err != nil {
		return nil, err
	}
	return out, nil
}

type copilotMCPProvenance struct {
	CanonicalSHA256 string            `json:"canonical_sha256"`
	Files           map[string]string `json:"files"`
}

func copilotMCP(base string, out map[string][]byte) error {
	canonicalPath := filepath.Join(base, "tools", "mcp.json")
	canonical, err := readSourceFile(canonicalPath)
	if errors.Is(err, fs.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	validated := map[string][]byte{}
	if err := mcpJSON(base, ".github/mcp.json", validated); err != nil {
		return err
	}
	provenanceData, err := readSourceFile(filepath.Join(base, "settings", "copilot-mcp-provenance.json"))
	if errors.Is(err, fs.ErrNotExist) {
		out[".github/mcp.json"] = canonical
		return nil
	}
	if err != nil {
		return err
	}
	var provenance copilotMCPProvenance
	if err := json.Unmarshal(provenanceData, &provenance); err != nil {
		return fmt.Errorf("settings/copilot-mcp-provenance.json: %w", err)
	}
	if len(provenance.Files) == 0 {
		return errors.New("settings/copilot-mcp-provenance.json must list at least one source file")
	}
	if provenance.CanonicalSHA256 == contentSHA256(canonical) {
		for path, raw := range provenance.Files {
			if !validCopilotMCPPath(path) {
				return fmt.Errorf("settings/copilot-mcp-provenance.json: invalid source path %q", path)
			}
			out[path] = []byte(raw)
		}
		return nil
	}
	if len(provenance.Files) != 1 {
		return errors.New("tools/mcp.json changed after importing both .mcp.json and .github/mcp.json; consolidate to one source before export")
	}
	for path := range provenance.Files {
		if !validCopilotMCPPath(path) {
			return fmt.Errorf("settings/copilot-mcp-provenance.json: invalid source path %q", path)
		}
		out[path] = canonical
	}
	return nil
}

func validCopilotMCPPath(path string) bool {
	return path == ".mcp.json" || path == ".github/mcp.json"
}

func contentSHA256(data []byte) string {
	return fmt.Sprintf("%x", sha256.Sum256(data))
}

func copilotAllowedModels(base string, out map[string][]byte) error {
	data, err := readSourceFile(filepath.Join(base, "permissions", "copilot-allowed-models.txt"))
	if errors.Is(err, fs.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	if len(strings.TrimSpace(string(data))) == 0 {
		return errors.New("permissions/copilot-allowed-models.txt must not be empty")
	}
	out[".github/allowed_models.txt"] = data
	return nil
}

func copilotSettings(base string, out map[string][]byte) error {
	data, err := readSourceFile(filepath.Join(base, "settings", "copilot.json"))
	if errors.Is(err, fs.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	var settings any
	if err := json.Unmarshal(data, &settings); err != nil {
		return fmt.Errorf("settings/copilot.json: %w", err)
	}
	if _, ok := settings.(map[string]any); !ok {
		return errors.New("settings/copilot.json must contain a JSON object")
	}
	if err := rejectInlineCredentials(settings, "settings/copilot.json"); err != nil {
		return err
	}
	out[".github/copilot/settings.json"] = data
	return nil
}

func copilotInstructions(base string, out map[string][]byte) error {
	files, err := markdownFiles(filepath.Join(base, "instructions"))
	if err != nil {
		return err
	}
	var sections []string
	var native []byte
	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			return err
		}
		if filepath.Base(file) == "AGENTS.md" {
			out["AGENTS.md"] = data
			continue
		}
		if filepath.Base(file) == "copilot-instructions.md" {
			native = data
			continue
		}
		if filepath.Base(file) == "CLAUDE.md" {
			continue
		}
		sections = append(sections, strings.TrimSpace(string(data)))
	}
	if len(native) > 0 && len(sections) == 0 {
		out[".github/copilot-instructions.md"] = native
		return nil
	}
	if len(native) > 0 {
		sections = append([]string{strings.TrimSpace(string(native))}, sections...)
	}
	if len(sections) > 0 {
		out[".github/copilot-instructions.md"] = []byte(strings.Join(sections, "\n\n") + "\n")
	}
	return nil
}

func copilotPrompts(base string, out map[string][]byte) error {
	files, err := markdownFiles(filepath.Join(base, "prompts"))
	if err != nil {
		return err
	}
	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			return err
		}
		name := strings.TrimSuffix(filepath.Base(file), ".md") + ".prompt.md"
		out[filepath.ToSlash(filepath.Join(".github", "prompts", name))] = data
	}
	return nil
}

func copyDirectory(source, destination string, out map[string][]byte) error {
	info, err := os.Lstat(source)
	if errors.Is(err, fs.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("refusing symlinked plugin directory %s", source)
	}
	if !info.IsDir() {
		return nil
	}
	return filepath.WalkDir(source, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			return nil
		}
		if entry.Type()&os.ModeSymlink != 0 {
			return fmt.Errorf("refusing symlinked plugin file %s", path)
		}
		relative, err := filepath.Rel(source, path)
		if err != nil {
			return err
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		destinationPath := filepath.ToSlash(filepath.Clean(filepath.Join(destination, relative)))
		if destinationPath == "." || destinationPath == "" || strings.HasPrefix(destinationPath, "../") {
			return fmt.Errorf("invalid projected path %q", destinationPath)
		}
		if _, exists := out[destinationPath]; exists {
			return fmt.Errorf("duplicate projected output path %q", destinationPath)
		}
		out[destinationPath] = data
		return nil
	})
}
