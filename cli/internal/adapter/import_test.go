package adapter

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestImportRoundTripCopilot(t *testing.T) {
	root := t.TempDir()
	write(t, root, ".github/copilot-instructions.md", "Use concise commits.\n")
	write(t, root, ".github/instructions/alpha.instructions.md", "Use gofmt.\n")
	write(t, root, ".github/agents/reviewer.agent.md", "Review carefully.\n")
	write(t, root, ".github/hooks/audit.json", `{"name":"audit","hooks":{}}`)
	write(t, root, ".github/mcp.json", `{"mcpServers":{"example":{"type":"stdio","command":"example"}}}`)

	a, err := NewForTarget(root, "copilot", false)
	if err != nil {
		t.Fatal(err)
	}

	plan, err := a.ImportPlan(false)
	if err != nil {
		t.Fatalf("ImportPlan() error = %v", err)
	}

	want := map[string]struct{}{
		".agents/instructions/copilot-instructions.md": {},
		".agents/rules/alpha.md":                       {},
		".agents/agents/reviewer.md":                   {},
		".agents/hooks/audit.json":                     {},
		".agents/tools/mcp.json":                       {},
	}
	if len(plan.Create) != len(want) {
		t.Fatalf("ImportPlan() create count = %d, want %d", len(plan.Create), len(want))
	}
	for _, path := range plan.Create {
		if _, ok := want[path]; !ok {
			t.Fatalf("unexpected path in import plan: %s", path)
		}
	}

	if err := a.Import(false); err != nil {
		t.Fatalf("Import() error = %v", err)
	}

	assertFileEquals(t, filepath.Join(root, ".agents/instructions/copilot-instructions.md"), "Use concise commits.\n")
	assertFileContains(t, filepath.Join(root, ".agents/rules/alpha.md"), "applyTo: \"**/*\"")
	assertFileContains(t, filepath.Join(root, ".agents/agents/reviewer.md"), "description: reviewer")
	assertFileContains(t, filepath.Join(root, ".agents/hooks/audit.json"), "\"name\":\"audit\"")
	assertFileContains(t, filepath.Join(root, ".agents/tools/mcp.json"), "\"mcpServers\":{\"example\"")
	assertFileContains(t, filepath.Join(root, ".agents/tools/mcp.json"), `"example"`)

	if _, err := os.Stat(filepath.Join(root, ".agents/tools/mcp.json")); err != nil {
		t.Fatalf("expected generated tool configuration: %v", err)
	}
}

func TestImportLegacyMigrationFixture(t *testing.T) {
	for _, target := range []string{"copilot", "codex", "claude"} {
		t.Run(target, func(t *testing.T) {
			root := t.TempDir()
			copyTree(t, filepath.Join("..", "..", "testdata", "shared", "legacy-migration"), root)
			a, err := NewForTarget(root, target, false)
			if err != nil {
				t.Fatal(err)
			}
			if err := a.Import(false); err != nil {
				t.Fatalf("Import() error = %v", err)
			}
			assertSnapshot(t, root, filepath.Join("..", "..", "testdata", "shared", "legacy-migration.expected", target))
		})
	}
}

func TestImportConflictingFilenamesAcrossSourceEntries(t *testing.T) {
	root := t.TempDir()
	copyTree(t, filepath.Join("..", "..", "testdata", "shared", "conflict"), root)

	a, err := NewForTarget(root, "copilot", false)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := a.ImportPlan(false); err == nil || !strings.Contains(err.Error(), "duplicate output path") {
		t.Fatalf("ImportPlan() error = %v, want duplicate output path", err)
	}
}

func TestImportRoundTripCodex(t *testing.T) {
	root := t.TempDir()
	write(t, root, "AGENTS.md", "Project conventions.\n")
	write(t, root, ".codex/agents/reviewer.toml", `name = "reviewer"
description = "review agent"
developer_instructions = "check PR diffs"
`)
	write(t, root, ".codex/config.toml", `[mcp_servers."default"]
type = "stdio"
command = "server"
args = ["--stdio"]
`)

	a, err := NewForTarget(root, "codex", false)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := a.ImportPlan(false); err != nil {
		t.Fatalf("ImportPlan() error = %v", err)
	}
	if err := a.Import(false); err != nil {
		t.Fatalf("Import() error = %v", err)
	}

	assertFileEquals(t, filepath.Join(root, ".agents/instructions/AGENTS.md"), "Project conventions.\n")
	assertFileContains(t, filepath.Join(root, ".agents/agents/reviewer.md"), "description: review agent")
	assertFileContains(t, filepath.Join(root, ".agents/tools/mcp.json"), `"default":{`)
}

func TestImportRoundTripClaude(t *testing.T) {
	root := t.TempDir()
	write(t, root, "CLAUDE.md", "Follow Claude conventions.\n")
	write(t, root, ".claude/rules/check.md", "---\napplyTo:\n  - \"**/*.go\"\n---\nRun golangci-lint.\n")
	write(t, root, ".claude/rules/format.md", "---\napplyTo: \"**/*.ts\"\n---\nRun format checks.\n")

	a, err := NewForTarget(root, "claude", false)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := a.ImportPlan(false); err != nil {
		t.Fatalf("ImportPlan() error = %v", err)
	}
	if err := a.Import(false); err != nil {
		t.Fatalf("Import() error = %v", err)
	}

	assertFileEquals(t, filepath.Join(root, ".agents/instructions/CLAUDE.md"), "Follow Claude conventions.\n")
	assertFileContains(t, filepath.Join(root, ".agents/rules/check.md"), "applyTo:")
	assertFileContains(t, filepath.Join(root, ".agents/rules/format.md"), "**/*.ts")
}

func TestImportPlanConflictHonorsForce(t *testing.T) {
	root := t.TempDir()
	write(t, root, ".github/copilot-instructions.md", "Use concise commits.\n")
	write(t, root, ".agents/instructions/copilot-instructions.md", "Project-owned override.\n")

	a, err := NewForTarget(root, "copilot", false)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := a.ImportPlan(false); err == nil || !strings.Contains(err.Error(), "import would overwrite existing") {
		t.Fatalf("ImportPlan() error = %v, want overwrite error", err)
	}

	plan, err := a.ImportPlan(true)
	if err != nil {
		t.Fatalf("ImportPlan(force) error = %v", err)
	}
	if len(plan.Update) != 1 || plan.Update[0] != ".agents/instructions/copilot-instructions.md" {
		t.Fatalf("ImportPlan(force) update = %v, want [%q]", plan.Update, ".agents/instructions/copilot-instructions.md")
	}
	if err := a.Import(true); err != nil {
		t.Fatalf("Import(force) error = %v", err)
	}
	assertFileEquals(t, filepath.Join(root, ".agents/instructions/copilot-instructions.md"), "Use concise commits.\n")
}

func TestImportCollisionForDuplicatePaths(t *testing.T) {
	outputs := map[string][]byte{}
	if err := addImportOutput(outputs, ".agents/rules/conflict.md", []byte("first")); err != nil {
		t.Fatal(err)
	}
	if err := addImportOutput(outputs, ".agents/rules/conflict.md", []byte("second")); err == nil || !strings.Contains(err.Error(), "duplicate output path") {
		t.Fatalf("addImportOutput() duplicate path error = %v", err)
	}
}

func assertFileEquals(t *testing.T, path, want string) {
	t.Helper()
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading %s: %v", path, err)
	}
	if string(got) != want {
		t.Fatalf("%s = %q, want %q", path, got, want)
	}
}

func assertFileContains(t *testing.T, path, want string) {
	t.Helper()
	g, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading %s: %v", path, err)
	}
	if !strings.Contains(string(g), want) {
		t.Fatalf("%s = %q, want contains %q", path, g, want)
	}

	if strings.HasSuffix(path, "tools/mcp.json") {
		var payload any
		if err := json.Unmarshal(g, &payload); err != nil {
			t.Fatalf("expected valid JSON in %s: %v", path, err)
		}
	}
}
