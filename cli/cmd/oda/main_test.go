package main

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Open-Dot-Agents/Open-Dot-Agents/cli/internal/adapter"

	"gopkg.in/yaml.v3"
)

func TestResolveTargetsAll(t *testing.T) {
	targets, err := resolveTargets("all")
	if err != nil {
		t.Fatal(err)
	}
	if len(targets) == 0 {
		t.Fatalf("resolveTargets(\"all\") returned no targets")
	}
	want := map[string]struct{}{
		"copilot": {},
		"codex":   {},
		"claude":  {},
	}
	for _, target := range targets {
		delete(want, target)
	}
	if len(want) > 0 {
		for target := range want {
			t.Fatalf("resolveTargets(\"all\") missing target %s", target)
		}
	}
}

func TestRunGenerateDryRunAndDiff(t *testing.T) {
	root := t.TempDir()
	instructionPath := filepath.Join(root, ".agents", "instructions", "team.md")
	if err := os.MkdirAll(filepath.Dir(instructionPath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(instructionPath, []byte("Use concise commits.\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	a, err := adapter.NewForTarget(root, "codex", false)
	if err != nil {
		t.Fatal(err)
	}
	result := run(runContext{dryRun: true, diff: true}, "generate", a)
	if result.Status != "ok" {
		t.Fatalf("run() status = %q, error = %q", result.Status, result.Error)
	}
	if len(result.Diff) == 0 {
		t.Fatal("expected diff output")
	}
	for _, line := range result.Diff {
		if !strings.HasPrefix(line, "A ") && !strings.HasPrefix(line, "M ") && !strings.HasPrefix(line, "D ") {
			t.Fatalf("unexpected diff line %q", line)
		}
	}
}

func TestRunGenerateDryRunWithCi(t *testing.T) {
	root := t.TempDir()
	instructionPath := filepath.Join(root, ".agents", "instructions", "team.md")
	if err := os.MkdirAll(filepath.Dir(instructionPath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(instructionPath, []byte("Use concise commits.\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	a, err := adapter.NewForTarget(root, "copilot", false)
	if err != nil {
		t.Fatal(err)
	}
	result := run(runContext{dryRun: true, ciMode: true}, "generate", a)
	if result.Status != "drift" {
		t.Fatalf("run(ci dry-run) status = %q, want drift", result.Status)
	}
	if result.Error == "" {
		t.Fatal("expected drift error text in CI dry-run result")
	}
}

func TestRunCheckDryRunWithCi(t *testing.T) {
	root := t.TempDir()
	instructionPath := filepath.Join(root, ".agents", "instructions", "team.md")
	if err := os.MkdirAll(filepath.Dir(instructionPath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(instructionPath, []byte("Use concise commits.\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	a, err := adapter.NewForTarget(root, "copilot", false)
	if err != nil {
		t.Fatal(err)
	}
	if err := a.Generate(false); err != nil {
		t.Fatalf("Generate() error = %v", err)
	}
	// mutate output to force check drift
	if err := os.WriteFile(filepath.Join(root, ".github", "copilot-instructions.md"), []byte("modified"), 0o644); err != nil {
		t.Fatal(err)
	}
	result := run(runContext{ciMode: true}, "check", a)
	if result.Status != "drift" {
		t.Fatalf("run(check ci) status = %q, want drift", result.Status)
	}
}

func TestRunValidateAllTargets(t *testing.T) {
	ctx := runContext{jsonOut: true}
	root := t.TempDir()
	seedValidationArtifactsForTests(t, root)
	if err := os.MkdirAll(filepath.Join(root, ".agents"), 0o755); err != nil {
		t.Fatal(err)
	}
	a, err := adapter.NewForTarget(root, "copilot", false)
	if err != nil {
		t.Fatal(err)
	}
	result := run(ctx, "validate", a)
	if result.Status != "ok" {
		t.Fatalf("run(validate) status = %q, err = %q", result.Status, result.Error)
	}
}

func TestRunValidateFailsOnStatusMismatch(t *testing.T) {
	root := t.TempDir()
	seedValidationArtifactsForTests(t, root)
	if err := mutateMappingsStatus(root, ".agents/mappings.yaml", ".agents", "rules", "codex", "supported"); err != nil {
		t.Fatal(err)
	}
	a, err := adapter.NewForTarget(root, "codex", false)
	if err != nil {
		t.Fatal(err)
	}
	result := run(runContext{}, "validate", a)
	if result.Status != "error" {
		t.Fatalf("run(validate) status = %q, err = %q", result.Status, result.Error)
	}
	if !strings.Contains(result.Error, "target codex status mismatch") {
		t.Fatalf("run(validate) error = %q, want status mismatch error", result.Error)
	}
}

func TestRunValidateFailsOnInvalidMappingStatusValue(t *testing.T) {
	root := t.TempDir()
	seedValidationArtifactsForTests(t, root)
	if err := mutateMappingsStatus(root, ".agents/mappings.yaml", ".agents", "rules", "codex", "unsupported-but-unused"); err != nil {
		t.Fatal(err)
	}
	a, err := adapter.NewForTarget(root, "copilot", false)
	if err != nil {
		t.Fatal(err)
	}
	result := run(runContext{}, "validate", a)
	if result.Status != "error" {
		t.Fatalf("run(validate) status = %q, err = %q", result.Status, result.Error)
	}
	if !strings.Contains(result.Error, "status \"unsupported-but-unused\" is invalid") {
		t.Fatalf("run(validate) error = %q, want invalid mapping status error", result.Error)
	}
}

func TestRunValidateAllTargetsWithOneFailure(t *testing.T) {
	root := t.TempDir()
	seedValidationArtifactsForTests(t, root)
	if err := mutateMappingsStatus(root, ".agents/mappings.yaml", ".agents", "rules", "codex", "supported"); err != nil {
		t.Fatal(err)
	}
	targets, err := resolveTargets("all")
	if err != nil {
		t.Fatal(err)
	}
	results := make(map[string]commandResult)
	var errorCount int
	for _, target := range targets {
		a, err := adapter.NewForTarget(root, target, false)
		if err != nil {
			t.Fatalf("adapter.NewForTarget(%q) error = %v", target, err)
		}
		result := run(runContext{}, "validate", a)
		results[target] = result
		if result.Status != "ok" {
			errorCount++
		}
	}
	if errorCount != 1 {
		t.Fatalf("validate all targets errorCount = %d, want 1", errorCount)
	}
	codexResult, ok := results["codex"]
	if !ok {
		t.Fatal("missing codex validation result")
	}
	if codexResult.Status != "error" || !strings.Contains(codexResult.Error, "status mismatch") || !strings.Contains(codexResult.Error, "target codex") {
		t.Fatalf("codex validate result = %#v, want codex status mismatch error", codexResult)
	}
	for target, result := range results {
		if target == "codex" {
			continue
		}
		if result.Status != "ok" {
			t.Fatalf("target %q validate status = %q, error = %q", target, result.Status, result.Error)
		}
	}
}

func TestRunValidateAllTargetsJSONPayload(t *testing.T) {
	root := t.TempDir()
	seedValidationArtifactsForTests(t, root)
	parsed, err := parseCLIArgs([]string{"--root", root, "validate", "--target", "all", "--format", "json"}, func() {})
	if err != nil {
		t.Fatalf("parseCLIArgs() error = %v", err)
	}
	targets, err := resolveTargets(parsed.target)
	if err != nil {
		t.Fatal(err)
	}
	payload := commandOutput{Command: parsed.command}
	for _, target := range targets {
		a, err := adapter.NewForTarget(root, target, false)
		if err != nil {
			t.Fatalf("adapter.NewForTarget(%q) error = %v", target, err)
		}
		result := run(runContext{jsonOut: true}, parsed.command, a)
		payload.Targets = append(payload.Targets, result)
	}
	data, err := json.Marshal(payload)
	if err != nil {
		t.Fatal(err)
	}
	var decoded commandOutput
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("json.Unmarshal(payload) error = %v", err)
	}
	if decoded.Command != "validate" {
		t.Fatalf("payload.command = %q, want validate", decoded.Command)
	}
	if len(decoded.Targets) != len(targets) {
		t.Fatalf("payload.targets = %d, want %d", len(decoded.Targets), len(targets))
	}
}

func TestRunValidateFailsWithMalformedRuleFrontMatter(t *testing.T) {
	root := t.TempDir()
	seedValidationArtifactsForTests(t, root)
	write(t, root, ".agents/rules/go.md", "run gofmt.\n")
	a, err := adapter.NewForTarget(root, "copilot", false)
	if err != nil {
		t.Fatal(err)
	}
	result := run(runContext{}, "validate", a)
	if result.Status != "error" {
		t.Fatalf("run(validate) status = %q, err = %q", result.Status, result.Error)
	}
	if !strings.Contains(result.Error, "must start with YAML front matter") {
		t.Fatalf("run(validate) error = %q, want front-matter parse error", result.Error)
	}
}

func TestRunValidateFailsWithFilePatternMismatch(t *testing.T) {
	root := t.TempDir()
	seedValidationArtifactsForTests(t, root)
	write(t, root, ".agents/rules/notes.txt", "apply this rule")
	a, err := adapter.NewForTarget(root, "copilot", false)
	if err != nil {
		t.Fatal(err)
	}
	result := run(runContext{}, "validate", a)
	if result.Status != "error" {
		t.Fatalf("run(validate) status = %q, err = %q", result.Status, result.Error)
	}
	if !strings.Contains(result.Error, "does not match allowed file patterns for \"rules\" category") {
		t.Fatalf("run(validate) error = %q, want file pattern mismatch error", result.Error)
	}
}

func TestRunValidateFailsWithMissingRequiredSchemaKey(t *testing.T) {
	root := t.TempDir()
	seedValidationArtifactsForTests(t, root)
	removeSchemaRequiredKey(t, root, "agents.schema.json", "canonical_root")
	write(t, root, ".agents/instructions/team.md", "Use concise commits.\n")
	a, err := adapter.NewForTarget(root, "copilot", false)
	if err != nil {
		t.Fatal(err)
	}
	result := run(runContext{}, "validate", a)
	if result.Status != "error" {
		t.Fatalf("run(validate) status = %q, err = %q", result.Status, result.Error)
	}
	if !strings.Contains(result.Error, "missing required top-level keys") {
		t.Fatalf("run(validate) error = %q, want schema required key error", result.Error)
	}
}

func TestRunValidateFailsOnMappingsManifestVersionMismatch(t *testing.T) {
	root := t.TempDir()
	seedValidationArtifactsForTests(t, root)
	if err := mutateMappingsManifestValue(root, ".agents/mappings.yaml", "format_version", "v0.0.2"); err != nil {
		t.Fatal(err)
	}
	write(t, root, ".agents/instructions/team.md", "Use concise commits.\n")
	a, err := adapter.NewForTarget(root, "copilot", false)
	if err != nil {
		t.Fatal(err)
	}
	result := run(runContext{}, "validate", a)
	if result.Status != "error" {
		t.Fatalf("run(validate) status = %q, err = %q", result.Status, result.Error)
	}
	if !strings.Contains(result.Error, "mappings manifest format_version must be \"v0.0.1\"") {
		t.Fatalf("run(validate) error = %q, want mappings manifest format_version error", result.Error)
	}
}

func TestRunValidateFailsOnMappingsCanonicalRootMismatch(t *testing.T) {
	root := t.TempDir()
	seedValidationArtifactsForTests(t, root)
	if err := mutateMappingsManifestValue(root, ".agents/mappings.yaml", "canonical_root", "docs"); err != nil {
		t.Fatal(err)
	}
	write(t, root, ".agents/instructions/team.md", "Use concise commits.\n")
	a, err := adapter.NewForTarget(root, "copilot", false)
	if err != nil {
		t.Fatal(err)
	}
	result := run(runContext{}, "validate", a)
	if result.Status != "error" {
		t.Fatalf("run(validate) status = %q, err = %q", result.Status, result.Error)
	}
	if !strings.Contains(result.Error, "mappings manifest canonical_root must be \".agents\"") {
		t.Fatalf("run(validate) error = %q, want mappings canonical_root error", result.Error)
	}
}

func TestTargetManifestDirectory(t *testing.T) {
	for _, target := range []string{"copilot", "codex", "claude"} {
		t.Run(target, func(t *testing.T) {
			manifestDir, err := targetManifestDirectory(target)
			if err != nil {
				t.Fatalf("targetManifestDirectory(%q) error = %v", target, err)
			}
			if manifestDir == "" {
				t.Fatalf("targetManifestDirectory(%q) returned empty manifest dir", target)
			}
		})
	}
	if _, err := targetManifestDirectory("does-not-exist"); err == nil {
		t.Fatal("targetManifestDirectory(\"does-not-exist\") expected error")
	}
}

func TestBackupGeneratedOutput(t *testing.T) {
	root := t.TempDir()
	source := filepath.Join(root, ".github")
	if err := os.MkdirAll(filepath.Join(source, "sub"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(source, "copilot-instructions.md"), []byte("old"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(source, "sub", "keep.txt"), []byte("keep"), 0o644); err != nil {
		t.Fatal(err)
	}

	backupRoot := filepath.Join(root, "backup")
	destination, err := backupGeneratedOutput(root, ".github", backupRoot)
	if err != nil {
		t.Fatal(err)
	}
	if destination == "" {
		t.Fatal("expected backup destination")
	}
	if _, err := os.Stat(destination); err != nil {
		t.Fatalf("backup destination not created: %v", err)
	}
	if data, err := os.ReadFile(filepath.Join(destination, "copilot-instructions.md")); err != nil || string(data) != "old" {
		t.Fatalf("backup contents missing/mismatched: data=%q err=%v", data, err)
	}
	if data, err := os.ReadFile(filepath.Join(destination, "sub", "keep.txt")); err != nil || string(data) != "keep" {
		t.Fatalf("backup contents missing/mismatched: data=%q err=%v", data, err)
	}
	if _, err := os.Stat(backupRoot); err != nil {
		t.Fatal(err)
	}
}

func TestRunGenerateWithBackup(t *testing.T) {
	root := t.TempDir()
	seedValidationArtifactsForTests(t, root)
	if err := os.MkdirAll(filepath.Join(root, ".agents", "instructions"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(root, ".github"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, ".agents", "instructions", "team.md"), []byte("Use concise commits.\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, ".github", "manual.txt"), []byte("manual"), 0o644); err != nil {
		t.Fatal(err)
	}

	a, err := adapter.NewForTarget(root, "copilot", false)
	if err != nil {
		t.Fatal(err)
	}
	backupRoot := filepath.Join(root, "runs-backups")
	result := run(runContext{root: root, force: true, backup: true, backupDir: backupRoot}, "generate", a)
	if result.Status != "ok" {
		t.Fatalf("run(generate) status = %q, err = %q", result.Status, result.Error)
	}
	if result.Backup == "" {
		t.Fatal("expected backup path in result")
	}
	if _, err := os.Stat(filepath.Join(result.Backup, "manual.txt")); err != nil {
		t.Fatalf("expected manual file in backup: %v", err)
	}
}

func TestParseCLIArgsSupportsCommandFirst(t *testing.T) {
	parsed, err := parseCLIArgs([]string{"validate", "--target", "all", "--dry-run"}, func() {})
	if err != nil {
		t.Fatalf("parseCLIArgs(command-first) error = %v", err)
	}
	if parsed.command != "validate" {
		t.Fatalf("parsed.command = %q, want validate", parsed.command)
	}
	if parsed.target != "all" {
		t.Fatalf("parsed.target = %q, want all", parsed.target)
	}
	if !parsed.dryRun {
		t.Fatal("parsed.dryRun = false, want true")
	}
}

func TestParseCLIArgsSupportsGuideCommand(t *testing.T) {
	parsed, err := parseCLIArgs([]string{"guide", "--target", "codex", "--format", "json"}, func() {})
	if err != nil {
		t.Fatalf("parseCLIArgs(guide command) error = %v", err)
	}
	if parsed.command != "guide" {
		t.Fatalf("parsed.command = %q, want guide", parsed.command)
	}
	if parsed.target != "codex" {
		t.Fatalf("parsed.target = %q, want codex", parsed.target)
	}
	if parsed.format != "json" {
		t.Fatalf("parsed.format = %q, want json", parsed.format)
	}
}

func TestParseCLIArgsSupportsFlagThenCommand(t *testing.T) {
	parsed, err := parseCLIArgs([]string{"--target", "all", "validate", "--dry-run"}, func() {})
	if err != nil {
		t.Fatalf("parseCLIArgs(flag-first) error = %v", err)
	}
	if parsed.command != "validate" {
		t.Fatalf("parsed.command = %q, want validate", parsed.command)
	}
	if parsed.target != "all" {
		t.Fatalf("parsed.target = %q, want all", parsed.target)
	}
	if !parsed.dryRun {
		t.Fatal("parsed.dryRun = false, want true")
	}
}

func TestParseCLIArgsSupportsGuideWithCommandFirstAndFlagAfterCommand(t *testing.T) {
	parsed, err := parseCLIArgs([]string{"guide", "--target", "codex", "--dry-run"}, func() {})
	if err != nil {
		t.Fatalf("parseCLIArgs(command-first + flag-after-command) error = %v", err)
	}
	if parsed.command != "guide" {
		t.Fatalf("parsed.command = %q, want guide", parsed.command)
	}
	if parsed.target != "codex" {
		t.Fatalf("parsed.target = %q, want codex", parsed.target)
	}
	if !parsed.dryRun {
		t.Fatal("parsed.dryRun = false, want true")
	}
}

func TestFindCommandArg(t *testing.T) {
	if idx := findCommandArg([]string{"--target", "all", "generate", "--dry-run"}); idx != 2 {
		t.Fatalf("findCommandArg(flag-first) = %d, want 2", idx)
	}
	if idx := findCommandArg([]string{"check", "--target", "copilot"}); idx != 0 {
		t.Fatalf("findCommandArg(command-first) = %d, want 0", idx)
	}
	if idx := findCommandArg([]string{"--help"}); idx != -1 {
		t.Fatalf("findCommandArg(help only) = %d, want -1", idx)
	}
}

func TestIsKnownCommandIncludesGuide(t *testing.T) {
	if !isKnownCommand("guide") {
		t.Fatal("isKnownCommand(\"guide\") = false, want true")
	}
	if isKnownCommand("unknown") {
		t.Fatal("isKnownCommand(\"unknown\") = true, want false")
	}
}

func TestCompletionScriptsIncludeGuideAndOptions(t *testing.T) {
	bashScript := bashCompletionScript()
	for _, token := range []string{"validate", "generate", "import", "check", "clean", "guide", "completion", "--help", "--target", "--format"} {
		if !strings.Contains(bashScript, token) {
			t.Fatalf("bashCompletionScript() missing %q", token)
		}
	}
	zshScript := zshCompletionScript()
	for _, token := range []string{"validate", "generate", "import", "check", "clean", "guide", "completion", "--help", "--target", "--format"} {
		if !strings.Contains(zshScript, token) {
			t.Fatalf("zshCompletionScript() missing %q", token)
		}
	}
}

func seedValidationArtifactsForTests(t *testing.T, root string) {
	t.Helper()
	sourceRoot := filepath.Join("..", "..", "..", ".agents")
	if err := copyArtifact(t, filepath.Join(sourceRoot, "schema", "v0.0.1", "agents.schema.json"), filepath.Join(root, ".agents", "schema", "v0.0.1", "agents.schema.json")); err != nil {
		t.Fatal(err)
	}
	if err := copyArtifact(t, filepath.Join(sourceRoot, "schema", "v0.0.1", "mappings.schema.json"), filepath.Join(root, ".agents", "schema", "v0.0.1", "mappings.schema.json")); err != nil {
		t.Fatal(err)
	}
	if err := copyArtifact(t, filepath.Join(sourceRoot, "mappings.yaml"), filepath.Join(root, ".agents", "mappings.yaml")); err != nil {
		t.Fatal(err)
	}
	if err := copyArtifact(t, filepath.Join(sourceRoot, "manifest.json"), filepath.Join(root, ".agents", "manifest.json")); err != nil {
		t.Fatal(err)
	}
}

func copyArtifact(t *testing.T, source, destination string) error {
	t.Helper()
	data, err := os.ReadFile(source)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(destination), 0o755); err != nil {
		return err
	}
	return os.WriteFile(destination, data, 0o644)
}

func mutateMappingsStatus(root, relativePath, rootKey, category, target, value string) error {
	data, err := os.ReadFile(filepath.Join(root, relativePath))
	if err != nil {
		return err
	}
	var mappings map[string]any
	if err := yaml.Unmarshal(data, &mappings); err != nil {
		return err
	}
	rootSection, ok := mappings[rootKey].(map[string]any)
	if !ok {
		return errors.New("mappings root section missing")
	}
	statusRaw, ok := rootSection["status"].(map[string]any)
	if !ok {
		return errors.New("mappings status section missing")
	}
	categoryStatuses, ok := statusRaw[category].(map[string]any)
	if !ok {
		return errors.New("mappings category status missing")
	}
	categoryStatuses[target] = value
	updated, err := yaml.Marshal(mappings)
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(root, relativePath), updated, 0o644)
}

func mutateMappingsManifestValue(root, relativePath, key string, value any) error {
	data, err := os.ReadFile(filepath.Join(root, relativePath))
	if err != nil {
		return err
	}
	var mappings map[string]any
	if err := yaml.Unmarshal(data, &mappings); err != nil {
		return err
	}
	rootSection, ok := mappings[".agents"].(map[string]any)
	if !ok {
		return errors.New("mappings root section missing")
	}
	manifest, ok := rootSection["manifest"].(map[string]any)
	if !ok {
		return errors.New("mappings manifest section missing")
	}
	manifest[key] = value
	updated, err := yaml.Marshal(mappings)
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(root, relativePath), updated, 0o644)
}

func removeSchemaRequiredKey(t *testing.T, root, filename, key string) {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(root, ".agents", "schema", "v0.0.1", filename))
	if err != nil {
		t.Fatal(err)
	}
	var doc map[string]any
	if err := json.Unmarshal(data, &doc); err != nil {
		t.Fatal(err)
	}
	rawRequired, ok := doc["required"].([]any)
	if !ok {
		t.Fatal("schema missing required key")
	}
	filtered := make([]any, 0, len(rawRequired))
	for _, required := range rawRequired {
		if required == key {
			continue
		}
		filtered = append(filtered, required)
	}
	doc["required"] = filtered
	updated, err := json.Marshal(doc)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, ".agents", "schema", "v0.0.1", filename), updated, 0o644); err != nil {
		t.Fatal(err)
	}
}

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
