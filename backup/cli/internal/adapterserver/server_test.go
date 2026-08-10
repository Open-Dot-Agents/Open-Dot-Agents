package adapterserver

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/Open-Dot-Agents/Open-Dot-Agents/cli/internal/adapter"
	"gopkg.in/yaml.v3"
)

func TestNormalizeImportPlanProducesV1CanonicalData(t *testing.T) {
	plan := &adapter.GenerationPlan{Outputs: map[string][]byte{
		".agents/rules/go.md":           []byte("---\nid: go\napplyTo: \"**/*.go, cmd/**\"\n---\nUse gofmt.\n"),
		".agents/tools/mcp.json":        []byte(`{"mcpServers":{"docs":{"type":"http","url":"https://example.com/mcp"}}}`),
		".agents/settings/copilot.json": []byte("{}\n"),
	}}

	if err := normalizeImportPlan(plan, "copilot"); err != nil {
		t.Fatal(err)
	}
	if _, exists := plan.Outputs[".agents/settings/copilot.json"]; exists {
		t.Fatal("vendor settings remained in the portable category")
	}
	if _, exists := plan.Outputs[".agents/extensions/org.open-dot-agents.copilot/settings/copilot.json"]; !exists {
		t.Fatal("vendor settings were not namespaced")
	}

	rule := string(plan.Outputs[".agents/rules/go.md"])
	frontEnd := strings.Index(rule[4:], "\n---\n")
	if frontEnd < 0 {
		t.Fatalf("invalid normalized rule: %q", rule)
	}
	var front map[string]any
	if err := yaml.Unmarshal([]byte(rule[4:4+frontEnd]), &front); err != nil {
		t.Fatal(err)
	}
	patterns, ok := front["applyTo"].([]any)
	if !ok || len(patterns) != 2 || patterns[0] != "**/*.go" || patterns[1] != "cmd/**" {
		t.Fatalf("applyTo = %#v", front["applyTo"])
	}

	var mcp struct {
		Servers map[string]map[string]any `json:"mcpServers"`
	}
	if err := json.Unmarshal(plan.Outputs[".agents/tools/mcp.json"], &mcp); err != nil {
		t.Fatal(err)
	}
	if got := mcp.Servers["docs"]["type"]; got != "streamable-http" {
		t.Fatalf("MCP transport = %v", got)
	}
}

func TestNormalizeImportPlanRejectsNormalizedCollision(t *testing.T) {
	plan := &adapter.GenerationPlan{Outputs: map[string][]byte{
		".agents/settings/a.json":                                      []byte("one"),
		".agents/extensions/org.open-dot-agents.codex/settings/a.json": []byte("two"),
	}}
	if err := normalizeImportPlan(plan, "codex"); err == nil || !strings.Contains(err.Error(), "collision") {
		t.Fatalf("error = %v", err)
	}
}

func TestConvertPortableHooksToClaude(t *testing.T) {
	raw := json.RawMessage(`[{"id":"test","event":"after-tool","matcher":{"tool":"Bash"},"action":{"type":"command","command":["task","test"]}}]`)
	data, losses, err := convertPortableHooksToClaude("default.json", raw)
	if err != nil {
		t.Fatal(err)
	}
	if len(losses) != 0 {
		t.Fatalf("losses = %#v", losses)
	}
	var document struct {
		Hooks map[string][]struct {
			Matcher string `json:"matcher"`
			Hooks   []struct {
				Type    string `json:"type"`
				Command string `json:"command"`
			} `json:"hooks"`
		} `json:"hooks"`
	}
	if err := json.Unmarshal(data, &document); err != nil {
		t.Fatal(err)
	}
	definitions := document.Hooks["PostToolUse"]
	if len(definitions) != 1 || definitions[0].Matcher != "Bash" || definitions[0].Hooks[0].Command != "'task' 'test'" {
		t.Fatalf("converted hooks = %s", data)
	}
}

func TestUnsupportedCategoryLossesAreDeterministic(t *testing.T) {
	root := t.TempDir()
	writeFixture := func(relative string) {
		t.Helper()
		path := filepath.Join(root, filepath.FromSlash(relative))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte("{}"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	writeFixture(".agents/settings/z.json")
	writeFixture(".agents/settings/a.json")
	losses, err := unsupportedCategoryLosses(root, map[string]string{"settings": "unsupported"})
	if err != nil {
		t.Fatal(err)
	}
	sort.Slice(losses, func(i, j int) bool { return losses[i].Path < losses[j].Path })
	if len(losses) != 2 || losses[0].Path != ".agents/settings/a.json" || losses[1].Path != ".agents/settings/z.json" {
		t.Fatalf("losses = %#v", losses)
	}
}
