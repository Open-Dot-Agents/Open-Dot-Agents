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
	write(t, root, "AGENTS.md", "Root instructions.\n")
	write(t, root, "services/api/AGENTS.md", "API instructions.\n")
	write(t, root, ".github/copilot-instructions.md", "Use concise commits.\n")
	write(t, root, "services/api/.github/copilot-instructions.md", "Nested Copilot instructions.\n")
	write(t, root, ".github/instructions/go/generated/alpha.instructions.md", "---\napplyTo: \"**/*.go\"\nunknown: retained\n---\nUse gofmt.\n")
	write(t, root, "services/api/.github/instructions/generated.instructions.md", "---\napplyTo: \"services/api/**\"\n---\nGenerated API guidance.\n")
	write(t, root, ".github/agents/reviewer.agent.md", "---\ndescription: Review carefully\ntools: [read]\n---\nReview carefully.\n")
	write(t, root, ".github/hooks/audit.json", `{"name":"audit","hooks":{}}`)
	write(t, root, ".github/mcp.json", `{"mcpServers":{"example":{"type":"stdio","command":"example"}}}`)
	write(t, root, ".github/copilot/settings.json", `{"enabledPlugins":{"review@team":true},"extraKnownMarketplaces":{"team":{"source":{"source":"github","repo":"example/plugins"}}}}`)
	write(t, root, ".github/allowed_models.txt", "fallback: gpt-5.6-terra\ngpt-5.6-*\n")
	write(t, root, ".github/skills/check/SKILL.md", "---\nname: check\ndescription: Check code\n---\nCheck it.\n")
	write(t, root, ".github/skills/check/scripts/run.sh", "#!/bin/sh\n")
	write(t, root, ".github/prompts/review.prompt.md", "Review ${selection}.\n")
	write(t, root, ".github/plugin/marketplace.json", `{"name":"local"}`)

	a, err := NewForTarget(root, "copilot", false)
	if err != nil {
		t.Fatal(err)
	}

	plan, err := a.ImportPlan(false)
	if err != nil {
		t.Fatalf("ImportPlan() error = %v", err)
	}

	want := map[string]struct{}{
		".agents/instructions/AGENTS.md":                                                            {},
		".agents/instructions/copilot-instructions.md":                                              {},
		".agents/instructions/copilot-project/services/api/AGENTS.md":                               {},
		".agents/instructions/copilot-project/services/api/.github/copilot-instructions.md":         {},
		".agents/rules/go/generated/alpha.md":                                                       {},
		".agents/rules/copilot-project/services/api/.github/instructions/generated.instructions.md": {},
		".agents/agents/reviewer.md":                                                                {},
		".agents/hooks/audit.json":                                                                  {},
		".agents/tools/mcp.json":                                                                    {},
		".agents/settings/copilot.json":                                                             {},
		".agents/permissions/copilot-allowed-models.txt":                                            {},
		".agents/skills/check/SKILL.md":                                                             {},
		".agents/skills/check/scripts/run.sh":                                                       {},
		".agents/prompts/review.md":                                                                 {},
		".agents/plugins/copilot/marketplace.json":                                                  {},
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
	seedSchemaArtifactsForValidation(t, root)
	if err := a.Validate(); err != nil {
		t.Fatalf("Validate(imported Copilot) error = %v", err)
	}

	assertFileEquals(t, filepath.Join(root, ".agents/instructions/copilot-instructions.md"), "Use concise commits.\n")
	assertFileEquals(t, filepath.Join(root, ".agents/rules/go/generated/alpha.md"), "---\napplyTo: \"**/*.go\"\nunknown: retained\n---\nUse gofmt.\n")
	assertFileEquals(t, filepath.Join(root, ".agents/agents/reviewer.md"), "---\ndescription: Review carefully\ntools: [read]\n---\nReview carefully.\n")
	assertFileContains(t, filepath.Join(root, ".agents/hooks/audit.json"), "\"name\":\"audit\"")
	assertFileContains(t, filepath.Join(root, ".agents/tools/mcp.json"), "\"mcpServers\":{\"example\"")
	assertFileContains(t, filepath.Join(root, ".agents/tools/mcp.json"), `"example"`)
	assertFileContains(t, filepath.Join(root, ".agents/settings/copilot.json"), `"review@team":true`)
	assertFileEquals(t, filepath.Join(root, ".agents/permissions/copilot-allowed-models.txt"), "fallback: gpt-5.6-terra\ngpt-5.6-*\n")

	if _, err := os.Stat(filepath.Join(root, ".agents/tools/mcp.json")); err != nil {
		t.Fatalf("expected generated tool configuration: %v", err)
	}
	if err := a.Generate(true); err != nil {
		t.Fatalf("Generate(force) error = %v", err)
	}
	assertFileEquals(t, filepath.Join(root, ".github/instructions/go/generated/alpha.instructions.md"), "---\napplyTo: \"**/*.go\"\nunknown: retained\n---\nUse gofmt.\n")
	assertFileEquals(t, filepath.Join(root, "services/api/AGENTS.md"), "API instructions.\n")
	assertFileEquals(t, filepath.Join(root, "services/api/.github/copilot-instructions.md"), "Nested Copilot instructions.\n")
	assertFileEquals(t, filepath.Join(root, "services/api/.github/instructions/generated.instructions.md"), "---\napplyTo: \"services/api/**\"\n---\nGenerated API guidance.\n")
	assertFileEquals(t, filepath.Join(root, ".github/agents/reviewer.agent.md"), "---\ndescription: Review carefully\ntools: [read]\n---\nReview carefully.\n")
	assertFileEquals(t, filepath.Join(root, ".github/skills/check/scripts/run.sh"), "#!/bin/sh\n")
	assertFileEquals(t, filepath.Join(root, ".github/prompts/review.prompt.md"), "Review ${selection}.\n")
	assertFileEquals(t, filepath.Join(root, "AGENTS.md"), "Root instructions.\n")
	assertFileEquals(t, filepath.Join(root, ".github/plugin/marketplace.json"), `{"name":"local"}`)
	assertFileEquals(t, filepath.Join(root, ".github/copilot/settings.json"), `{"enabledPlugins":{"review@team":true},"extraKnownMarketplaces":{"team":{"source":{"source":"github","repo":"example/plugins"}}}}`)
	assertFileEquals(t, filepath.Join(root, ".github/allowed_models.txt"), "fallback: gpt-5.6-terra\ngpt-5.6-*\n")
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

func TestImportCopilotRootMCPPreservesSourceLocation(t *testing.T) {
	root := t.TempDir()
	original := "{\n  \"mcpServers\": {\"root\": {\"type\": \"stdio\", \"command\": \"root-server\"}}\n}\n"
	write(t, root, ".mcp.json", original)
	a, err := NewForTarget(root, "copilot", false)
	if err != nil {
		t.Fatal(err)
	}
	if err := a.Import(false); err != nil {
		t.Fatal(err)
	}
	assertFileContains(t, filepath.Join(root, ".agents/settings/copilot-mcp-provenance.json"), `".mcp.json"`)
	if err := a.Generate(true); err != nil {
		t.Fatal(err)
	}
	assertFileEquals(t, filepath.Join(root, ".mcp.json"), original)
	if _, err := os.Stat(filepath.Join(root, ".github/mcp.json")); !os.IsNotExist(err) {
		t.Fatalf("unexpected .github/mcp.json: %v", err)
	}
}

func TestImportCopilotDualMCPPreservesBothSources(t *testing.T) {
	root := t.TempDir()
	rootMCP := `{"mcpServers":{"root":{"type":"stdio","command":"root-server"}}}`
	githubMCP := "{\n  \"mcpServers\": {\"github\": {\"type\": \"stdio\", \"command\": \"github-server\"}}\n}\n"
	write(t, root, ".mcp.json", rootMCP)
	write(t, root, ".github/mcp.json", githubMCP)
	a, err := NewForTarget(root, "copilot", false)
	if err != nil {
		t.Fatal(err)
	}
	if err := a.Import(false); err != nil {
		t.Fatal(err)
	}
	assertFileContains(t, filepath.Join(root, ".agents/tools/mcp.json"), `"root"`)
	assertFileContains(t, filepath.Join(root, ".agents/tools/mcp.json"), `"github"`)
	if err := a.Generate(true); err != nil {
		t.Fatal(err)
	}
	assertFileEquals(t, filepath.Join(root, ".mcp.json"), rootMCP)
	assertFileEquals(t, filepath.Join(root, ".github/mcp.json"), githubMCP)

	write(t, root, ".agents/tools/mcp.json", `{"mcpServers":{"changed":{"type":"stdio","command":"changed"}}}`)
	if err := a.Generate(true); err == nil || !strings.Contains(err.Error(), "consolidate to one source") {
		t.Fatalf("Generate() error = %v, want ambiguous dual-source diagnostic", err)
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
	agentTOML := `# retained agent comment
name = "reviewer"
description = "review agent"
developer_instructions = "check PR diffs"
model = "gpt-5.4"
[metadata]
unknown = true
`
	write(t, root, ".codex/agents/reviewer.toml", agentTOML)
	configTOML := `# retained config comment
model = "gpt-5.4"
approval_policy = "on-request"
sandbox_mode = "workspace-write"
unknown_future_field = 42
project_doc_fallback_filenames = ["TEAM_GUIDE.md"]

[profiles.review]
model_reasoning_effort = "high"

[mcp_servers."default"]
command = "server"
args = ["--stdio"]
startup_timeout_sec = 15
`
	write(t, root, ".codex/config.toml", configTOML)
	write(t, root, "AGENTS.override.md", "Root override.\n")
	write(t, root, "services/payments/AGENTS.override.md", "Payment override.\n")
	write(t, root, "services/payments/TEAM_GUIDE.md", "Payment fallback.\n")
	write(t, root, ".codex/rules/team/git.rules", "prefix_rule(pattern = [\"git\", \"status\"], decision = \"allow\")\n")
	write(t, root, ".codex/skills/legacy/SKILL.md", "---\nname: legacy\ndescription: Legacy skill\n---\nUse it.\n")

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
	seedSchemaArtifactsForValidation(t, root)
	if err := a.Validate(); err != nil {
		t.Fatalf("Validate(imported Codex) error = %v", err)
	}

	assertFileEquals(t, filepath.Join(root, ".agents/instructions/AGENTS.md"), "Project conventions.\n")
	assertFileContains(t, filepath.Join(root, ".agents/agents/reviewer.md"), `name: "reviewer"`)
	assertFileContains(t, filepath.Join(root, ".agents/agents/reviewer.md"), `description: "review agent"`)
	assertFileContains(t, filepath.Join(root, ".agents/tools/mcp.json"), `"default": {`)
	assertFileContains(t, filepath.Join(root, ".agents/tools/mcp.json"), `"startup_timeout_sec": 15`)
	assertFileContains(t, filepath.Join(root, ".agents/settings/codex.toml"), "unknown_future_field = 42")
	assertFileEquals(t, filepath.Join(root, ".agents/instructions/codex-project/AGENTS.override.md"), "Root override.\n")
	assertFileEquals(t, filepath.Join(root, ".agents/instructions/codex-project/services/payments/AGENTS.override.md"), "Payment override.\n")
	assertFileEquals(t, filepath.Join(root, ".agents/instructions/codex-project/services/payments/TEAM_GUIDE.md"), "Payment fallback.\n")
	assertFileContains(t, filepath.Join(root, ".agents/permissions/codex.toml"), `approval_policy = 'on-request'`)
	assertFileEquals(t, filepath.Join(root, ".agents/permissions/codex-rules/team/git.rules"), "prefix_rule(pattern = [\"git\", \"status\"], decision = \"allow\")\n")
	assertFileContains(t, filepath.Join(root, ".agents/profiles/codex.toml"), "[profiles]")
	assertFileContains(t, filepath.Join(root, ".agents/settings/codex-legacy-skills.json"), `"legacy/SKILL.md"`)
	if err := a.Generate(true); err != nil {
		t.Fatalf("Generate(force) error = %v", err)
	}
	assertFileEquals(t, filepath.Join(root, ".codex/agents/reviewer.toml"), agentTOML)
	assertFileEquals(t, filepath.Join(root, ".codex/config.toml"), configTOML)
	assertFileEquals(t, filepath.Join(root, "AGENTS.override.md"), "Root override.\n")
	assertFileEquals(t, filepath.Join(root, "services/payments/AGENTS.override.md"), "Payment override.\n")
	assertFileEquals(t, filepath.Join(root, "services/payments/TEAM_GUIDE.md"), "Payment fallback.\n")
	assertFileEquals(t, filepath.Join(root, ".codex/rules/team/git.rules"), "prefix_rule(pattern = [\"git\", \"status\"], decision = \"allow\")\n")
	assertFileEquals(t, filepath.Join(root, ".codex/skills/legacy/SKILL.md"), "---\nname: legacy\ndescription: Legacy skill\n---\nUse it.\n")
}

func TestCanonicalRoundTripThroughCopilot(t *testing.T) {
	source := t.TempDir()
	write(t, source, ".agents/instructions/AGENTS.md", "Root instructions.\n")
	write(t, source, ".agents/instructions/copilot-project/services/api/AGENTS.md", "API instructions.\n")
	write(t, source, ".agents/instructions/copilot-project/services/api/.github/copilot-instructions.md", "Nested Copilot instructions.\n")
	write(t, source, ".agents/instructions/copilot-instructions.md", "Copilot instructions.\n")
	write(t, source, ".agents/rules/go.md", "---\napplyTo: \"**/*.go\"\n---\nRun gofmt.\n")
	write(t, source, ".agents/rules/copilot-project/services/api/.github/instructions/api.instructions.md", "---\napplyTo: \"services/api/**\"\n---\nAPI guidance.\n")
	write(t, source, ".agents/agents/reviewer.md", "---\ndescription: Review changes\n---\nReview.\n")
	write(t, source, ".agents/prompts/review.md", "Review ${selection}.\n")
	write(t, source, ".agents/plugins/copilot/marketplace.json", `{"name":"local"}`)
	write(t, source, ".agents/settings/copilot.json", `{"enabledPlugins":{"review@team":true}}`)
	write(t, source, ".agents/permissions/copilot-allowed-models.txt", "fallback: gpt-5.6-terra\ngpt-5.6-*\n")
	write(t, source, ".agents/skills/check/SKILL.md", "---\nname: check\ndescription: Check code\n---\nCheck.\n")
	a, err := NewForTarget(source, "copilot", false)
	if err != nil {
		t.Fatal(err)
	}
	if err := a.Generate(false); err != nil {
		t.Fatal(err)
	}

	fresh := t.TempDir()
	copyTree(t, filepath.Join(source, ".github"), filepath.Join(fresh, ".github"))
	copyTree(t, filepath.Join(source, "services"), filepath.Join(fresh, "services"))
	copyFileForRoundTrip(t, filepath.Join(source, "AGENTS.md"), filepath.Join(fresh, "AGENTS.md"))
	imported, err := NewForTarget(fresh, "copilot", false)
	if err != nil {
		t.Fatal(err)
	}
	if err := imported.Import(false); err != nil {
		t.Fatal(err)
	}
	if err := imported.Generate(true); err != nil {
		t.Fatal(err)
	}
	assertTreesEqual(t, filepath.Join(source, ".github"), filepath.Join(fresh, ".github"), true)
	assertTreesEqual(t, filepath.Join(source, "services"), filepath.Join(fresh, "services"), false)
	assertFileEquals(t, filepath.Join(fresh, "AGENTS.md"), "Root instructions.\n")
}

func TestCanonicalRoundTripThroughCodex(t *testing.T) {
	source := t.TempDir()
	write(t, source, ".agents/instructions/AGENTS.md", "Codex instructions.\n")
	write(t, source, ".agents/instructions/codex-project/services/payments/AGENTS.override.md", "Payment override.\n")
	write(t, source, ".agents/agents/reviewer.md", "---\ndescription: Review changes\n---\nReview.\n")
	write(t, source, ".agents/settings/codex.toml", "model = 'gpt-5.4'\nfuture = 7\n")
	write(t, source, ".agents/permissions/codex.toml", "approval_policy = 'on-request'\n")
	write(t, source, ".agents/permissions/codex-rules/team/git.rules", "prefix_rule(pattern = [\"git\", \"status\"], decision = \"allow\")\n")
	write(t, source, ".agents/profiles/codex.toml", "[profiles.review]\nmodel_reasoning_effort = 'high'\n")
	write(t, source, ".agents/tools/mcp.json", `{"mcpServers":{"docs":{"type":"http","url":"https://developers.openai.com/mcp","codex":{"startup_timeout_sec":15}}}}`)
	write(t, source, ".agents/skills/check/SKILL.md", "---\nname: check\ndescription: Check code\n---\nCheck.\n")
	a, err := NewForTarget(source, "codex", false)
	if err != nil {
		t.Fatal(err)
	}
	if err := a.Generate(false); err != nil {
		t.Fatal(err)
	}

	fresh := t.TempDir()
	copyTree(t, filepath.Join(source, ".codex"), filepath.Join(fresh, ".codex"))
	copyTree(t, filepath.Join(source, "services"), filepath.Join(fresh, "services"))
	copyTree(t, filepath.Join(source, ".agents", "skills"), filepath.Join(fresh, ".agents", "skills"))
	copyFileForRoundTrip(t, filepath.Join(source, "AGENTS.md"), filepath.Join(fresh, "AGENTS.md"))
	imported, err := NewForTarget(fresh, "codex", false)
	if err != nil {
		t.Fatal(err)
	}
	if err := imported.Import(false); err != nil {
		t.Fatal(err)
	}
	if err := imported.Generate(true); err != nil {
		t.Fatal(err)
	}
	assertTreesEqual(t, filepath.Join(source, ".codex"), filepath.Join(fresh, ".codex"), true)
	assertTreesEqual(t, filepath.Join(source, "services"), filepath.Join(fresh, "services"), false)
	assertFileEquals(t, filepath.Join(fresh, "AGENTS.md"), "Codex instructions.\n")
}

func copyFileForRoundTrip(t *testing.T, source, destination string) {
	t.Helper()
	data, err := os.ReadFile(source)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Dir(destination), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(destination, data, 0o644); err != nil {
		t.Fatal(err)
	}
}

func assertTreesEqual(t *testing.T, wantRoot, gotRoot string, ignoreManifest bool) {
	t.Helper()
	err := filepath.WalkDir(wantRoot, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil || entry.IsDir() {
			return walkErr
		}
		if ignoreManifest && entry.Name() == manifestName {
			return nil
		}
		relative, err := filepath.Rel(wantRoot, path)
		if err != nil {
			return err
		}
		want, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		got, err := os.ReadFile(filepath.Join(gotRoot, relative))
		if err != nil {
			return err
		}
		if string(want) != string(got) {
			t.Errorf("round-trip mismatch for %s", filepath.ToSlash(relative))
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
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
	if _, err := a.ImportPlan(false); err == nil || !strings.Contains(err.Error(), "not adapter-owned") {
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

func TestImportManifestTracksOwnedUpdatesAndStaleDeletes(t *testing.T) {
	root := t.TempDir()
	write(t, root, ".github/copilot-instructions.md", "first\n")
	a, err := NewForTarget(root, "copilot", false)
	if err != nil {
		t.Fatal(err)
	}
	if err := a.Import(false); err != nil {
		t.Fatal(err)
	}
	assertFileContains(t, filepath.Join(root, ".agents/.open-dot-agents-import-copilot.json"), `"target": "copilot"`)

	write(t, root, ".github/copilot-instructions.md", "second\n")
	plan, err := a.ImportPlan(false)
	if err != nil {
		t.Fatalf("owned update: %v", err)
	}
	if len(plan.Update) != 1 || plan.Update[0] != ".agents/instructions/copilot-instructions.md" {
		t.Fatalf("owned update plan = %#v", plan)
	}
	if err := a.Import(false); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(filepath.Join(root, ".github/copilot-instructions.md")); err != nil {
		t.Fatal(err)
	}
	plan, err = a.ImportPlan(false)
	if err != nil {
		t.Fatalf("stale delete: %v", err)
	}
	if len(plan.Delete) != 1 || plan.Delete[0] != ".agents/instructions/copilot-instructions.md" {
		t.Fatalf("stale delete plan = %#v", plan)
	}
	if err := a.Import(false); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(root, ".agents/instructions/copilot-instructions.md")); !os.IsNotExist(err) {
		t.Fatalf("stale imported file still exists: %v", err)
	}
}

func TestImportManifestProtectsModifiedOwnedCanonicalFile(t *testing.T) {
	root := t.TempDir()
	write(t, root, ".github/copilot-instructions.md", "vendor\n")
	a, err := NewForTarget(root, "copilot", false)
	if err != nil {
		t.Fatal(err)
	}
	if err := a.Import(false); err != nil {
		t.Fatal(err)
	}
	write(t, root, ".agents/instructions/copilot-instructions.md", "canonical edit\n")
	if _, err := a.ImportPlan(false); err == nil || !strings.Contains(err.Error(), "was modified") {
		t.Fatalf("ImportPlan() error = %v, want modified owned-file diagnostic", err)
	}
	if err := a.Import(true); err != nil {
		t.Fatalf("Import(force) error = %v", err)
	}
	assertFileEquals(t, filepath.Join(root, ".agents/instructions/copilot-instructions.md"), "vendor\n")
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

func TestImportRejectsInlineRepositoryCredentials(t *testing.T) {
	for _, test := range []struct {
		name   string
		target string
		path   string
		data   string
	}{
		{"copilot-mcp", "copilot", ".github/mcp.json", `{"mcpServers":{"remote":{"type":"http","url":"https://example.test","headers":{"Authorization":"Bearer literal-secret"}}}}`},
		{"copilot-settings", "copilot", ".github/copilot/settings.json", `{"accessToken":"literal-secret"}`},
		{"codex", "codex", ".codex/config.toml", "[mcp_servers.remote]\nurl = 'https://example.test'\n[mcp_servers.remote.http_headers]\nAuthorization = 'Bearer literal-secret'\n"},
	} {
		t.Run(test.name, func(t *testing.T) {
			root := t.TempDir()
			write(t, root, test.path, test.data)
			a, err := NewForTarget(root, test.target, false)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := a.ImportPlan(false); err == nil || !strings.Contains(err.Error(), "inline credential") {
				t.Fatalf("ImportPlan() error = %v, want inline credential diagnostic", err)
			}
		})
	}
}

func TestImportCopilotIgnoresPersonalLocalSettings(t *testing.T) {
	root := t.TempDir()
	write(t, root, ".github/copilot/settings.local.json", `{"accessToken":"personal-local-value"}`)
	a, err := NewForTarget(root, "copilot", false)
	if err != nil {
		t.Fatal(err)
	}
	plan, err := a.ImportPlan(false)
	if err != nil {
		t.Fatalf("ImportPlan() error = %v", err)
	}
	if len(plan.Create) != 0 || len(plan.Update) != 0 || len(plan.Delete) != 0 {
		t.Fatalf("ImportPlan() imported local settings: %#v", plan)
	}
}

func TestImportRejectsSymlinkedRepositoryArtifact(t *testing.T) {
	root := t.TempDir()
	external := filepath.Join(t.TempDir(), "SKILL.md")
	if err := os.WriteFile(external, []byte("outside\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(root, ".github", "skills", "outside", "SKILL.md")
	if err := os.MkdirAll(filepath.Dir(link), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(external, link); err != nil {
		t.Fatal(err)
	}
	a, err := NewForTarget(root, "copilot", false)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := a.ImportPlan(false); err == nil || !strings.Contains(err.Error(), "symlinked import source") {
		t.Fatalf("ImportPlan() error = %v, want symlink diagnostic", err)
	}
}

func TestCodexExportRejectsOrphanedAgentSidecar(t *testing.T) {
	root := t.TempDir()
	write(t, root, ".agents/settings/codex-agents/orphan.toml", "name = 'orphan'\n")
	a, err := NewForTarget(root, "codex", false)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := a.Plan(false); err == nil || !strings.Contains(err.Error(), "no matching canonical agent") {
		t.Fatalf("Plan() error = %v, want orphaned sidecar diagnostic", err)
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
