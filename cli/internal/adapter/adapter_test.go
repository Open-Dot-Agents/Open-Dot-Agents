package adapter

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func write(t *testing.T, root, path, content string) {
	t.Helper()
	full := filepath.Join(root, path)
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(full, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestGenerateCheckAndClean(t *testing.T) {
	root := t.TempDir()
	write(t, root, ".agents/instructions/team.md", "Use concise commits.\n")
	write(t, root, ".agents/rules/go.md", "---\napplyTo: \"**/*.go\"\n---\nRun gofmt.\n")
	write(t, root, ".agents/agents/reviewer.md", "---\ndescription: Reviews changes\n---\nReview carefully.\n")
	write(t, root, ".agents/hooks/audit.json", `{"version":1,"hooks":{"sessionStart":[]}}`)
	write(t, root, ".agents/tools/mcp.json", `{"mcpServers":{"example":{"type":"stdio","command":"example"}}}`)
	write(t, root, ".agents/skills/release/SKILL.md", "---\nname: release\ndescription: Prepare a release\n---\nFollow the release process.\n")

	a := New(root, false)
	if err := a.Generate(false); err != nil {
		t.Fatalf("Generate() error = %v", err)
	}
	for _, path := range []string{
		".github/copilot-instructions.md",
		".github/instructions/go.instructions.md",
		".github/agents/reviewer.agent.md",
		".github/hooks/audit.json",
		".github/mcp.json",
		".github/.open-dot-agents.json",
	} {
		if _, err := os.Stat(filepath.Join(root, path)); err != nil {
			t.Errorf("generated %s: %v", path, err)
		}
	}
	if err := a.Check(); err != nil {
		t.Fatalf("Check() error = %v", err)
	}
	if err := a.Generate(false); err != nil {
		t.Fatalf("second Generate() error = %v", err)
	}
	if err := a.Clean(); err != nil {
		t.Fatalf("Clean() error = %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, ".github/mcp.json")); !os.IsNotExist(err) {
		t.Fatalf("mcp output remains after clean: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, ".github")); !os.IsNotExist(err) {
		t.Fatalf("empty compatibility directory remains after clean: %v", err)
	}
}

func TestRejectsUnsupportedCategory(t *testing.T) {
	root := t.TempDir()
	seedSchemaArtifactsForValidation(t, root)
	write(t, root, ".agents/prompts/review.md", "Review this.\n")

	err := New(root, false).Validate()
	if err == nil || !strings.Contains(err.Error(), "prompts") {
		t.Fatalf("Validate() error = %v, want unsupported prompts", err)
	}
	if err := New(root, true).Validate(); err != nil {
		t.Fatalf("Validate() with acknowledgement error = %v", err)
	}
}

func TestRefusesUnownedCompatibilityTree(t *testing.T) {
	root := t.TempDir()
	write(t, root, ".agents/tools/mcp.json", `{"mcpServers":{}}`)
	write(t, root, ".github/mcp.json", `{"mcpServers":{"manual":{}}}`)

	err := New(root, false).Generate(false)
	if err == nil || !strings.Contains(err.Error(), "not adapter-owned") {
		t.Fatalf("Generate() error = %v, want ownership error", err)
	}
}

func TestRuleRequiresApplyTo(t *testing.T) {
	root := t.TempDir()
	seedSchemaArtifactsForValidation(t, root)
	write(t, root, ".agents/rules/missing.md", "---\nname: missing\n---\nNo scope.\n")

	err := New(root, false).Validate()
	if err == nil || !strings.Contains(err.Error(), "requires applyTo") {
		t.Fatalf("Validate() error = %v, want applyTo error", err)
	}
}

func TestRejectsUnsafeManifestPath(t *testing.T) {
	root := t.TempDir()
	write(t, root, ".github/.open-dot-agents.json", `{"version":1,"files":{"../outside":"hash"}}`)

	err := New(root, false).Clean()
	if err == nil || !strings.Contains(err.Error(), "invalid adapter manifest path") {
		t.Fatalf("Clean() error = %v, want unsafe manifest path error", err)
	}
}

func TestCopilotFixtureSnapshot(t *testing.T) {
	runSharedFixture(t, "copilot", "basic", false, ".github/copilot-instructions.md")
}

func TestCodexGenerateCheckAndClean(t *testing.T) {
	runSharedFixture(t, "codex", "basic", false, "AGENTS.md")
}

func TestCodexUsesCanonicalSkillsWithoutProjection(t *testing.T) {
	root := t.TempDir()
	write(t, root, ".agents/skills/release/SKILL.md", "---\nname: release\ndescription: Prepare a release\n---\nReturn ODA_SKILL_OK.\n")

	a, err := NewForTarget(root, "codex", false)
	if err != nil {
		t.Fatal(err)
	}
	plan, err := a.Plan(false)
	if err != nil {
		t.Fatalf("Plan() error = %v", err)
	}
	for path := range plan.Outputs {
		if strings.HasPrefix(path, ".codex/skills/") {
			t.Fatalf("Plan() projected duplicate Codex skill %q", path)
		}
	}
	if err := a.Generate(false); err != nil {
		t.Fatalf("Generate() error = %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, ".agents/skills/release/SKILL.md")); err != nil {
		t.Fatalf("canonical skill missing after generation: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, ".codex/skills")); !os.IsNotExist(err) {
		t.Fatalf("Codex skill projection exists: %v", err)
	}
}

func TestClaudeGenerateCheckAndClean(t *testing.T) {
	runSharedFixture(t, "claude", "basic", false, "CLAUDE.md")
}

func TestCopilotComplexFixture(t *testing.T) {
	runSharedFixture(t, "copilot", "complex", false, ".github/copilot-instructions.md")
}

func TestCodexComplexFixture(t *testing.T) {
	root := t.TempDir()
	seedSchemaArtifactsForValidation(t, root)
	copyTree(t, filepath.Join("..", "..", "testdata", "shared", "complex"), root)
	a, err := NewForTarget(root, "codex", false)
	if err != nil {
		t.Fatal(err)
	}
	if err := a.Validate(); err == nil ||
		!strings.Contains(err.Error(), "unsupported populated categories") ||
		!strings.Contains(err.Error(), "hooks") ||
		!strings.Contains(err.Error(), "rules") {
		t.Fatalf("Validate() error = %v, want unsupported hooks and rules", err)
	}
	runSharedFixture(t, "codex", "complex", true, "AGENTS.md")
}

func TestClaudeComplexFixture(t *testing.T) {
	runSharedFixture(t, "claude", "complex", false, "CLAUDE.md")
}

func TestMalformedSharedFixtureValidation(t *testing.T) {
	root := t.TempDir()
	seedSchemaArtifactsForValidation(t, root)
	copyTree(t, filepath.Join("..", "..", "testdata", "shared", "malformed"), root)
	a, err := NewForTarget(root, "copilot", false)
	if err != nil {
		t.Fatal(err)
	}
	if err := a.Validate(); err == nil || !strings.Contains(err.Error(), "must start with YAML front matter") {
		t.Fatalf("Validate() error = %v, want front matter diagnostic", err)
	}
}

func TestEmptyOptionalCategoryFixtures(t *testing.T) {
	for _, test := range []struct {
		target      string
		allowUnsafe bool
		anchor      string
	}{
		{"copilot", false, ".github/copilot-instructions.md"},
		{"codex", false, "AGENTS.md"},
		{"claude", false, "CLAUDE.md"},
	} {
		t.Run(test.target, func(t *testing.T) {
			runSharedFixture(t, test.target, "empty-optional", test.allowUnsafe, test.anchor)
		})
	}
}

func TestNestedSkillAssetsSharedFixture(t *testing.T) {
	for _, test := range []struct {
		target string
		anchor string
		force  bool
	}{
		{"copilot", ".github/copilot-instructions.md", false},
		{"codex", "AGENTS.md", false},
		{"claude", "CLAUDE.md", false},
	} {
		t.Run(test.target, func(t *testing.T) {
			runSharedFixture(t, test.target, "nested-skills", test.force, test.anchor)
		})
	}
}

func TestRejectsUnknownTarget(t *testing.T) {
	if _, err := NewForTarget(t.TempDir(), "unknown", false); err == nil {
		t.Fatal("NewForTarget() accepted an unknown target")
	}
}

func TestRendererContracts(t *testing.T) {
	tests := []struct {
		target      string
		manifestDir string
		unsupported []string
	}{
		{"copilot", ".github", commonUnsupportedCategories},
		{"codex", ".codex", append(append([]string{}, commonUnsupportedCategories...), "hooks", "rules")},
		{"claude", ".claude", commonUnsupportedCategories},
	}
	for _, test := range tests {
		t.Run(test.target, func(t *testing.T) {
			renderer, err := rendererFor(test.target)
			if err != nil {
				t.Fatal(err)
			}
			if renderer.target() != test.target {
				t.Errorf("target() = %q, want %q", renderer.target(), test.target)
			}
			if renderer.manifestDirectory() != test.manifestDir {
				t.Errorf("manifestDirectory() = %q, want %q", renderer.manifestDirectory(), test.manifestDir)
			}
			if strings.Join(renderer.unsupportedCategories(), ",") != strings.Join(test.unsupported, ",") {
				t.Errorf("unsupportedCategories() = %v, want %v", renderer.unsupportedCategories(), test.unsupported)
			}
		})
	}
}

func TestRejectsInvalidSharedConfiguration(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		content string
		want    string
	}{
		{
			name:    "non-string rule applyTo",
			path:    ".agents/rules/go.md",
			content: "---\napplyTo:\n  - \"**/*.go\"\n---\nRun gofmt.\n",
			want:    "requires applyTo",
		},
		{
			name:    "non-string agent description",
			path:    ".agents/agents/reviewer.md",
			content: "---\ndescription: 42\n---\nReview carefully.\n",
			want:    "requires description",
		},
		{
			name:    "non-string skill metadata",
			path:    ".agents/skills/release/SKILL.md",
			content: "---\nname: release\ndescription: 42\n---\nRelease.\n",
			want:    "requires name and description",
		},
		{
			name:    "MCP server is not object",
			path:    ".agents/tools/mcp.json",
			content: `{"mcpServers":{"invalid":"server"}}`,
			want:    "must be an object",
		},
		{
			name:    "MCP server name is empty",
			path:    ".agents/tools/mcp.json",
			content: `{"mcpServers":{"":{"type":"stdio","command":"server"}}}`,
			want:    "empty MCP server name",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			root := t.TempDir()
			seedSchemaArtifactsForValidation(t, root)
			write(t, root, test.path, test.content)
			err := New(root, false).Validate()
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("Validate() error = %v, want %q", err, test.want)
			}
		})
	}
}

func TestCodexRejectsUnsupportedAndInvalidMCP(t *testing.T) {
	t.Run("rules are unsupported", func(t *testing.T) {
		root := t.TempDir()
		seedSchemaArtifactsForValidation(t, root)
		write(t, root, ".agents/rules/go.md", "---\napplyTo: \"**/*.go\"\n---\nRun gofmt.\n")
		a, err := NewForTarget(root, "codex", false)
		if err != nil {
			t.Fatal(err)
		}
		if err := a.Validate(); err == nil || !strings.Contains(err.Error(), "rules") {
			t.Fatalf("Validate() error = %v, want unsupported rules", err)
		}
	})
	for _, test := range []struct {
		name, config, want string
	}{
		{"stdio needs command", `{"mcpServers":{"server":{"type":"stdio"}}}`, "requires command"},
		{"HTTP needs URL", `{"mcpServers":{"server":{"type":"http"}}}`, "requires url"},
		{"SSE unsupported", `{"mcpServers":{"server":{"type":"sse","url":"https://example.test"}}}`, "unsupported Codex transport"},
	} {
		t.Run(test.name, func(t *testing.T) {
			root := t.TempDir()
			seedSchemaArtifactsForValidation(t, root)
			write(t, root, ".agents/tools/mcp.json", test.config)
			a, err := NewForTarget(root, "codex", false)
			if err != nil {
				t.Fatal(err)
			}
			if err := a.Validate(); err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("Validate() error = %v, want %q", err, test.want)
			}
		})
	}
}

func TestCodexMCPTransportVariants(t *testing.T) {
	for _, test := range []struct {
		name   string
		config string
		want   string
	}{
		{"local transport", `{"mcpServers":{"local":{"type":"local","command":"server","args":["--stdio"]}}}`, "[mcp_servers.\"local\"]\ncommand = \"server\"\nargs = [\"--stdio\"]\n"},
		{"streamable http transport", `{"mcpServers":{"remote":{"type":"streamable-http","url":"https://example.test/mcp","headers":{"Authorization":"token"}}}}`, "[mcp_servers.\"remote\"]\nurl = \"https://example.test/mcp\"\n"},
	} {
		t.Run(test.name, func(t *testing.T) {
			root := t.TempDir()
			seedSchemaArtifactsForValidation(t, root)
			write(t, root, ".agents/tools/mcp.json", test.config)
			a, err := NewForTarget(root, "codex", false)
			if err != nil {
				t.Fatal(err)
			}
			if err := a.Generate(false); err != nil {
				t.Fatalf("Generate() error = %v", err)
			}
			got, err := os.ReadFile(filepath.Join(root, ".codex/config.toml"))
			if err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(string(got), strings.TrimSuffix(test.want, "\n")) {
				t.Fatalf("generated config = %q, want includes %q", got, test.want)
			}
		})
	}
}

func TestCodexRejectsMalformedMCPTransportConfig(t *testing.T) {
	for _, test := range []struct {
		name   string
		config string
		want   string
	}{
		{"unsupported transport", `{"mcpServers":{"server":{"type":"sse","url":"https://example.test"}}}`, "unsupported Codex transport"},
		{"http missing url", `{"mcpServers":{"server":{"type":"http"}}}`, "requires url"},
		{"streamable-http missing url", `{"mcpServers":{"server":{"type":"streamable-http"}}}`, "requires url"},
		{"local missing command", `{"mcpServers":{"server":{"type":"local"}}}`, "requires command"},
	} {
		t.Run(test.name, func(t *testing.T) {
			root := t.TempDir()
			seedSchemaArtifactsForValidation(t, root)
			write(t, root, ".agents/tools/mcp.json", test.config)
			a, err := NewForTarget(root, "codex", false)
			if err != nil {
				t.Fatal(err)
			}
			if err := a.Validate(); err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("Validate() error = %v, want %q", err, test.want)
			}
		})
	}
}

func TestCodexDefaultStdioMCPTransport(t *testing.T) {
	root := t.TempDir()
	seedSchemaArtifactsForValidation(t, root)
	write(t, root, ".agents/tools/mcp.json", `{"mcpServers":{"default":{"command":"server","args":["--stdio"]}}}`)
	a, err := NewForTarget(root, "codex", false)
	if err != nil {
		t.Fatal(err)
	}
	if err := a.Generate(false); err != nil {
		t.Fatalf("Generate() error = %v", err)
	}
	data, err := os.ReadFile(filepath.Join(root, ".codex/config.toml"))
	if err != nil {
		t.Fatal(err)
	}
	want := "#:schema " + codexConfigSchemaURL + "\n\n[mcp_servers.\"default\"]\ncommand = \"server\"\nargs = [\"--stdio\"]\n"
	if string(data) != want {
		t.Fatalf("Codex default stdio config = %q, want %q", data, want)
	}
}

func TestRejectsUnsafeGeneratedOutputPath(t *testing.T) {
	root := t.TempDir()
	seedSchemaArtifactsForValidation(t, root)
	write(t, root, ".agents/agents/reviewer.md", "---\nname: ../../../outside\ndescription: Reviews changes\n---\nReview carefully.\n")
	a, err := NewForTarget(root, "codex", false)
	if err != nil {
		t.Fatal(err)
	}
	if err := a.Validate(); err == nil || !strings.Contains(err.Error(), "unsafe generated output path") {
		t.Fatalf("Validate() error = %v, want unsafe generated path error", err)
	}
	if _, err := os.Stat(filepath.Join(root, "outside.toml")); !os.IsNotExist(err) {
		t.Fatalf("unsafe output was created: %v", err)
	}
}

func TestGeneratePlan(t *testing.T) {
	root := t.TempDir()
	write(t, root, ".agents/instructions/team.md", "Use concise commits.\n")
	a, err := NewForTarget(root, "copilot", false)
	if err != nil {
		t.Fatal(err)
	}
	plan, err := a.Plan(false)
	if err != nil {
		t.Fatalf("Plan() error = %v", err)
	}
	if len(plan.Create) != 1 || plan.Create[0] != ".github/copilot-instructions.md" {
		t.Fatalf("Plan() create = %v, want [%q]", plan.Create, ".github/copilot-instructions.md")
	}
}

func TestForceAllowsUnownedOverwriteInPlan(t *testing.T) {
	root := t.TempDir()
	write(t, root, ".agents/instructions/team.md", "Use concise commits.\n")
	write(t, root, ".github/copilot-instructions.md", "Existing manual output.\n")
	a, err := NewForTarget(root, "copilot", false)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := a.Plan(false); err == nil {
		t.Fatal("Plan() expected unowned collision without --force")
	}
	plan, err := a.Plan(true)
	if err != nil {
		t.Fatalf("Plan(force) error = %v", err)
	}
	if len(plan.Update) != 1 || plan.Update[0] != ".github/copilot-instructions.md" {
		t.Fatalf("Plan(force) update = %v, want [%q]", plan.Update, ".github/copilot-instructions.md")
	}
}

func TestEmptyMCPConfiguration(t *testing.T) {
	for _, test := range []struct {
		target string
		output string
		exists bool
	}{
		{"copilot", ".github/mcp.json", true},
		{"codex", ".codex/config.toml", false},
		{"claude", ".mcp.json", true},
	} {
		t.Run(test.target, func(t *testing.T) {
			root := t.TempDir()
			write(t, root, ".agents/tools/mcp.json", `{"mcpServers":{}}`)
			a, err := NewForTarget(root, test.target, false)
			if err != nil {
				t.Fatal(err)
			}
			if err := a.Generate(false); err != nil {
				t.Fatalf("Generate() error = %v", err)
			}
			_, err = os.Stat(filepath.Join(root, test.output))
			if test.exists && err != nil {
				t.Fatalf("expected generated %s: %v", test.output, err)
			}
			if !test.exists && !os.IsNotExist(err) {
				t.Fatalf("unexpected generated %s: %v", test.output, err)
			}
		})
	}
}

func TestClaudeRejectsDuplicateHookEvent(t *testing.T) {
	root := t.TempDir()
	seedSchemaArtifactsForValidation(t, root)
	write(t, root, ".agents/hooks/one.json", `{"hooks":{"SessionStart":[]}}`)
	write(t, root, ".agents/hooks/two.json", `{"hooks":{"SessionStart":[]}}`)
	a, err := NewForTarget(root, "claude", false)
	if err != nil {
		t.Fatal(err)
	}
	if err := a.Validate(); err == nil || !strings.Contains(err.Error(), "more than one file") {
		t.Fatalf("Validate() error = %v, want duplicate hook error", err)
	}
}

func TestClaudeRejectsDuplicateRulePathsAfterSplit(t *testing.T) {
	root := t.TempDir()
	seedSchemaArtifactsForValidation(t, root)
	write(t, root, ".agents/rules/backend.md", "---\napplyTo: \"api/**/*.go,internal/**/*.go\"\n---\nRun API checks.\n")
	write(t, root, ".agents/rules/frontend.md", "---\napplyTo: \"internal/**/*.go,web/**/*.js\"\n---\nRun frontend checks.\n")
	a, err := NewForTarget(root, "claude", false)
	if err != nil {
		t.Fatal(err)
	}
	if err := a.Validate(); err == nil || !strings.Contains(err.Error(), `path pattern "internal/**/*.go"`) {
		t.Fatalf("Validate() error = %v, want duplicate path pattern", err)
	}
}

func TestSplitGlobPatterns(t *testing.T) {
	got := splitGlobPatterns("web/**/*.{ts,tsx}, api/**/*.go, scripts/*.{js,ts}")
	want := []string{"web/**/*.{ts,tsx}", "api/**/*.go", "scripts/*.{js,ts}"}
	if strings.Join(got, "|") != strings.Join(want, "|") {
		t.Fatalf("splitGlobPatterns() = %v, want %v", got, want)
	}
}

func TestRejectsMalformedHooksAndSkills(t *testing.T) {
	tests := []struct {
		name    string
		target  string
		path    string
		content string
		want    string
	}{
		{
			name:    "Copilot invalid hook JSON",
			target:  "copilot",
			path:    ".agents/hooks/audit.json",
			content: "{",
			want:    "not valid JSON",
		},
		{
			name:    "Claude hook lacks hooks object",
			target:  "claude",
			path:    ".agents/hooks/audit.json",
			content: `{"version":1}`,
			want:    "must contain a hooks object",
		},
		{
			name:    "skill entrypoint is missing",
			target:  "copilot",
			path:    ".agents/skills/release/note.md",
			content: "Release notes.\n",
			want:    "entry \"skills/release/note.md\" does not match allowed file patterns for",
		},
		{
			name:    "skill front matter is malformed",
			target:  "claude",
			path:    ".agents/skills/release/SKILL.md",
			content: "No front matter.\n",
			want:    "requires name and description",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			root := t.TempDir()
			seedSchemaArtifactsForValidation(t, root)
			write(t, root, test.path, test.content)
			a, err := NewForTarget(root, test.target, false)
			if err != nil {
				t.Fatal(err)
			}
			if err := a.Validate(); err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("Validate() error = %v, want %q", err, test.want)
			}
		})
	}
}

func TestTargetCollisionAndUnsupportedAcknowledgement(t *testing.T) {
	for _, test := range []struct {
		target string
		output string
	}{
		{"copilot", ".github/copilot-instructions.md"},
		{"codex", "AGENTS.md"},
		{"claude", "CLAUDE.md"},
	} {
		t.Run(test.target+" collision", func(t *testing.T) {
			root := t.TempDir()
			write(t, root, ".agents/instructions/team.md", "Use project conventions.\n")
			write(t, root, test.output, "Manual configuration.\n")
			a, err := NewForTarget(root, test.target, false)
			if err != nil {
				t.Fatal(err)
			}
			if err := a.Generate(false); err == nil || !strings.Contains(err.Error(), "not adapter-owned") {
				t.Fatalf("Generate() error = %v, want ownership error", err)
			}
		})
	}
	for _, target := range []string{"copilot", "codex", "claude"} {
		t.Run(target+" unsupported acknowledgement", func(t *testing.T) {
			root := t.TempDir()
			seedSchemaArtifactsForValidation(t, root)
			write(t, root, ".agents/prompts/review.md", "Review changes.\n")
			a, err := NewForTarget(root, target, true)
			if err != nil {
				t.Fatal(err)
			}
			if err := a.Validate(); err != nil {
				t.Fatalf("Validate() with acknowledgement error = %v", err)
			}
		})
	}
}

func TestEmptyConfigurationProducesNoOutput(t *testing.T) {
	for _, target := range []string{"copilot", "codex", "claude"} {
		t.Run(target, func(t *testing.T) {
			root := t.TempDir()
			seedSchemaArtifactsForValidation(t, root)
			if err := os.MkdirAll(filepath.Join(root, ".agents"), 0o755); err != nil {
				t.Fatal(err)
			}
			a, err := NewForTarget(root, target, false)
			if err != nil {
				t.Fatal(err)
			}
			if err := a.Validate(); err != nil {
				t.Fatalf("Validate() error = %v", err)
			}
			if err := a.Generate(false); err != nil {
				t.Fatalf("Generate() error = %v", err)
			}
			if err := a.Check(); err != nil {
				t.Fatalf("Check() error = %v", err)
			}
			if err := a.Clean(); err != nil {
				t.Fatalf("Clean() error = %v", err)
			}
		})
	}
}

func TestTargetLifecycleGuards(t *testing.T) {
	tests := []struct {
		target string
		output string
	}{
		{"copilot", ".github/copilot-instructions.md"},
		{"codex", "AGENTS.md"},
		{"claude", "CLAUDE.md"},
	}
	for _, test := range tests {
		t.Run(test.target, func(t *testing.T) {
			root := t.TempDir()
			write(t, root, ".agents/instructions/team.md", "Use project conventions.\n")
			a, err := NewForTarget(root, test.target, false)
			if err != nil {
				t.Fatal(err)
			}
			if err := a.Generate(false); err != nil {
				t.Fatalf("Generate() error = %v", err)
			}
			write(t, root, test.output, "Manually modified.\n")
			if err := a.Check(); err == nil || !strings.Contains(err.Error(), "stale or modified") {
				t.Fatalf("Check() error = %v, want stale output", err)
			}
			if err := a.Clean(); err == nil || !strings.Contains(err.Error(), "refusing to remove") {
				t.Fatalf("Clean() error = %v, want modified output refusal", err)
			}
			if err := a.Generate(true); err != nil {
				t.Fatalf("Generate(force) error = %v", err)
			}
			if err := a.Check(); err != nil {
				t.Fatalf("Check() after forced generation error = %v", err)
			}
			if err := a.Clean(); err != nil {
				t.Fatalf("Clean() after forced generation error = %v", err)
			}
			if _, err := os.Stat(filepath.Join(root, test.output)); !os.IsNotExist(err) {
				t.Fatalf("generated output remains after clean: %v", err)
			}
		})
	}
}

func TestGenerationRemovesObsoleteOutput(t *testing.T) {
	tests := []struct {
		target string
		source string
		output string
	}{
		{"copilot", ".agents/agents/reviewer.md", ".github/agents/reviewer.agent.md"},
		{"codex", ".agents/agents/reviewer.md", ".codex/agents/reviewer.toml"},
		{"claude", ".agents/agents/reviewer.md", ".claude/agents/reviewer.md"},
	}
	for _, test := range tests {
		t.Run(test.target, func(t *testing.T) {
			root := t.TempDir()
			write(t, root, test.source, "---\ndescription: Reviews changes\n---\nReview carefully.\n")
			a, err := NewForTarget(root, test.target, false)
			if err != nil {
				t.Fatal(err)
			}
			if err := a.Generate(false); err != nil {
				t.Fatalf("Generate() error = %v", err)
			}
			if err := os.Remove(filepath.Join(root, test.source)); err != nil {
				t.Fatal(err)
			}
			if err := a.Generate(false); err != nil {
				t.Fatalf("Generate() after source removal error = %v", err)
			}
			if _, err := os.Stat(filepath.Join(root, test.output)); !os.IsNotExist(err) {
				t.Fatalf("obsolete generated output remains: %v", err)
			}
			if err := a.Check(); err != nil {
				t.Fatalf("Check() after source removal error = %v", err)
			}
		})
	}
}

func copyTree(t *testing.T, source, destination string) {
	t.Helper()
	err := filepath.WalkDir(source, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		relative, err := filepath.Rel(source, path)
		if err != nil {
			return err
		}
		target := filepath.Join(destination, relative)
		if entry.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		return os.WriteFile(target, data, 0o644)
	})
	if err != nil {
		t.Fatal(err)
	}
}

func seedSchemaArtifactsForValidation(t *testing.T, root string) {
	t.Helper()

	sourceRoot := filepath.Join("..", "..", "..", ".agents")
	copyTree(t, filepath.Join(sourceRoot, "schema"), filepath.Join(root, ".agents", "schema"))
	copyFile(t, filepath.Join(sourceRoot, "mappings.yaml"), filepath.Join(root, ".agents", "mappings.yaml"))
	copyFile(t, filepath.Join(sourceRoot, "manifest.json"), filepath.Join(root, ".agents", "manifest.json"))
}

func copyFile(t *testing.T, source, destination string) {
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

func runSharedFixture(t *testing.T, target, fixture string, allowUnsupported bool, generatedAnchor string) {
	t.Helper()
	source := filepath.Join("..", "..", "testdata", "shared", fixture)
	expected := filepath.Join("..", "..", "testdata", "shared", fixture+".expected", target)
	root := t.TempDir()
	copyTree(t, source, root)
	a, err := NewForTarget(root, target, allowUnsupported)
	if err != nil {
		t.Fatal(err)
	}
	if err := a.Generate(false); err != nil {
		t.Fatalf("Generate() error = %v", err)
	}
	assertSnapshot(t, root, expected)
	if err := a.Check(); err != nil {
		t.Fatalf("Check() error = %v", err)
	}
	if err := a.Clean(); err != nil {
		t.Fatalf("Clean() error = %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, generatedAnchor)); !os.IsNotExist(err) {
		t.Fatalf("generated output remains after clean: %v", err)
	}
}

func assertSnapshot(t *testing.T, root, expectedRoot string) {
	t.Helper()
	err := filepath.WalkDir(expectedRoot, func(path string, entry os.DirEntry, err error) error {
		if err != nil || entry.IsDir() {
			return err
		}
		relative, err := filepath.Rel(expectedRoot, path)
		if err != nil {
			return err
		}
		want, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		got, err := os.ReadFile(filepath.Join(root, relative))
		if err != nil {
			return err
		}
		if string(got) != string(want) {
			t.Errorf("snapshot mismatch for %s", relative)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
