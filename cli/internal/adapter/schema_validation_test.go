package adapter

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestValidateRejectsMissingSchemaRequiredTopLevel(t *testing.T) {
	root := t.TempDir()
	seedSchemaArtifactsForValidation(t, root)
	if err := removeSchemaRequiredKey(root, "agents.schema.json", "canonical_root"); err != nil {
		t.Fatal(err)
	}
	write(t, root, ".agents/instructions/team.md", "Use concise commits.\n")

	a, err := NewForTarget(root, "copilot", false)
	if err != nil {
		t.Fatal(err)
	}
	if err := a.Validate(); err == nil || !strings.Contains(err.Error(), "missing required top-level keys") {
		t.Fatalf("Validate() error = %v, want missing schema required key error", err)
	}
}

func TestValidateRejectsInvalidMappingsStatusValue(t *testing.T) {
	root := t.TempDir()
	seedSchemaArtifactsForValidation(t, root)
	if err := mutateMappingsStatus(root, ".agents/mappings.yaml", ".agents", "rules", "copilot", "unknown"); err != nil {
		t.Fatal(err)
	}
	write(t, root, ".agents/rules/go.md", "---\napplyTo: \"**/*.go\"\n---\nRun gofmt.\n")

	a, err := NewForTarget(root, "copilot", false)
	if err != nil {
		t.Fatal(err)
	}
	if err := a.Validate(); err == nil || !strings.Contains(err.Error(), "status \"unknown\" is invalid") {
		t.Fatalf("Validate() error = %v, want invalid mapping status error", err)
	}
}

func TestValidateRejectsMappingsStatusUnknownTarget(t *testing.T) {
	root := t.TempDir()
	seedSchemaArtifactsForValidation(t, root)
	if err := mutateMappingsStatusTarget(root, ".agents/mappings.yaml", ".agents", "rules", "ghost", "supported"); err != nil {
		t.Fatal(err)
	}
	a, err := NewForTarget(root, "copilot", false)
	if err != nil {
		t.Fatal(err)
	}
	if err := a.Validate(); err == nil || !strings.Contains(err.Error(), "mappings references unknown target \"ghost\" for category \"rules\"") {
		t.Fatalf("Validate() error = %v, want unknown target in status error", err)
	}
}

func TestValidateRejectsTargetMappingStatusMismatch(t *testing.T) {
	root := t.TempDir()
	seedSchemaArtifactsForValidation(t, root)
	if err := mutateMappingsStatus(root, ".agents/mappings.yaml", ".agents", "rules", "codex", "supported"); err != nil {
		t.Fatal(err)
	}
	a, err := NewForTarget(root, "codex", false)
	if err != nil {
		t.Fatal(err)
	}
	if err := a.Validate(); err == nil || !strings.Contains(err.Error(), "target codex status mismatch") {
		t.Fatalf("Validate() error = %v, want mappings/renderer status mismatch error", err)
	}
}

func TestValidateRejectsMappingsMissingCategoryStatus(t *testing.T) {
	root := t.TempDir()
	seedSchemaArtifactsForValidation(t, root)
	if err := removeMappingsStatusCategory(root, ".agents/mappings.yaml", ".agents", "hooks"); err != nil {
		t.Fatal(err)
	}
	a, err := NewForTarget(root, "copilot", false)
	if err != nil {
		t.Fatal(err)
	}
	if err := a.Validate(); err == nil || !strings.Contains(err.Error(), "mappings missing status for category \"hooks\"") {
		t.Fatalf("Validate() error = %v, want missing category status error", err)
	}
}

func TestValidateRejectsMappingsMissingTargetStatusEntries(t *testing.T) {
	root := t.TempDir()
	seedSchemaArtifactsForValidation(t, root)
	if err := removeMappingsTargetFromStatus(root, ".agents/mappings.yaml", ".agents", "codex"); err != nil {
		t.Fatal(err)
	}
	a, err := NewForTarget(root, "codex", false)
	if err != nil {
		t.Fatal(err)
	}
	if err := a.Validate(); err == nil || !strings.Contains(err.Error(), "mappings target \"codex\" is missing required category status entries") {
		t.Fatalf("Validate() error = %v, want missing target status entries error", err)
	}
}

func TestValidateRejectsMappingsManifestFormatVersionMismatch(t *testing.T) {
	root := t.TempDir()
	seedSchemaArtifactsForValidation(t, root)
	if err := mutateMappingsManifestValue(root, ".agents/mappings.yaml", "format_version", "v0.0.2"); err != nil {
		t.Fatal(err)
	}
	write(t, root, ".agents/instructions/team.md", "Use concise commits.\n")

	a, err := NewForTarget(root, "copilot", false)
	if err != nil {
		t.Fatal(err)
	}
	if err := a.Validate(); err == nil || !strings.Contains(err.Error(), "mappings manifest format_version must be \"v0.0.1\"") {
		t.Fatalf("Validate() error = %v, want mappings manifest format mismatch error", err)
	}
}

func TestValidateRejectsMappingsCanonicalRootMismatch(t *testing.T) {
	root := t.TempDir()
	seedSchemaArtifactsForValidation(t, root)
	if err := mutateMappingsManifestValue(root, ".agents/mappings.yaml", "canonical_root", "docs"); err != nil {
		t.Fatal(err)
	}
	write(t, root, ".agents/instructions/team.md", "Use concise commits.\n")

	a, err := NewForTarget(root, "copilot", false)
	if err != nil {
		t.Fatal(err)
	}
	if err := a.Validate(); err == nil || !strings.Contains(err.Error(), "mappings manifest canonical_root must be \".agents\"") {
		t.Fatalf("Validate() error = %v, want mappings canonical-root mismatch error", err)
	}
}

func TestValidateRejectsManifestCompatibilityMismatch(t *testing.T) {
	root := t.TempDir()
	seedSchemaArtifactsForValidation(t, root)
	if err := mutateManifestValue(root, ".agents/manifest.json", "format_version", "v0.0.2"); err != nil {
		t.Fatal(err)
	}
	write(t, root, ".agents/instructions/team.md", "Use concise commits.\n")

	a, err := NewForTarget(root, "copilot", false)
	if err != nil {
		t.Fatal(err)
	}
	if err := a.Validate(); err == nil || !strings.Contains(err.Error(), "format_version == \"v0.0.1\"") {
		t.Fatalf("Validate() error = %v, want manifest format_version mismatch error", err)
	}
}

func TestValidateRejectsManifestCanonicalRootMismatch(t *testing.T) {
	root := t.TempDir()
	seedSchemaArtifactsForValidation(t, root)
	if err := mutateManifestValue(root, ".agents/manifest.json", "canonical_root", "docs"); err != nil {
		t.Fatal(err)
	}
	write(t, root, ".agents/instructions/team.md", "Use concise commits.\n")

	a, err := NewForTarget(root, "copilot", false)
	if err != nil {
		t.Fatal(err)
	}
	if err := a.Validate(); err == nil || !strings.Contains(err.Error(), "canonical_root == \".agents\"") {
		t.Fatalf("Validate() error = %v, want manifest canonical_root mismatch error", err)
	}
}

func TestValidateRejectsUnsupportedFilePatternForKnownCategory(t *testing.T) {
	root := t.TempDir()
	seedSchemaArtifactsForValidation(t, root)
	write(t, root, ".agents/rules/notes.txt", "apply this rule")

	adapter, err := NewForTarget(root, "copilot", false)
	if err != nil {
		t.Fatal(err)
	}
	if err := adapter.Validate(); err == nil || !strings.Contains(err.Error(), "does not match allowed file patterns for \"rules\" category") {
		t.Fatalf("Validate() error = %v, want file pattern mismatch error", err)
	}
}

func TestValidateRejectsMissingAgentsSchema(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, ".agents"), 0o755); err != nil {
		t.Fatal(err)
	}
	a, err := NewForTarget(root, "copilot", false)
	if err != nil {
		t.Fatal(err)
	}
	if err := a.Validate(); err == nil || !strings.Contains(err.Error(), "schema not available: .agents/schema/v0.0.1/agents.schema.json") {
		t.Fatalf("Validate() error = %v, want agents schema unavailable error", err)
	}
}

func TestValidateRejectsMissingMappingsFile(t *testing.T) {
	root := t.TempDir()
	seedSchemaArtifactsForValidation(t, root)
	if err := os.Remove(filepath.Join(root, ".agents", "mappings.yaml")); err != nil {
		t.Fatal(err)
	}
	a, err := NewForTarget(root, "copilot", false)
	if err != nil {
		t.Fatal(err)
	}
	if err := a.Validate(); err == nil || !strings.Contains(err.Error(), "schema not available: .agents/mappings.yaml") {
		t.Fatalf("Validate() error = %v, want mappings file unavailable error", err)
	}
}

func TestValidateAllowsMissingManifestCompatibilityFile(t *testing.T) {
	root := t.TempDir()
	seedSchemaArtifactsForValidation(t, root)
	if err := os.Remove(filepath.Join(root, ".agents", "manifest.json")); err != nil {
		t.Fatal(err)
	}
	write(t, root, ".agents/instructions/team.md", "Use concise commits.\n")

	a, err := NewForTarget(root, "copilot", false)
	if err != nil {
		t.Fatal(err)
	}
	if err := a.Validate(); err != nil {
		t.Fatalf("Validate() error = %v, want nil when manifest compatibility file is missing", err)
	}
}

func TestGenerateAllowsMissingManifestCompatibilityFile(t *testing.T) {
	root := t.TempDir()
	seedSchemaArtifactsForValidation(t, root)
	if err := os.Remove(filepath.Join(root, ".agents", "manifest.json")); err != nil {
		t.Fatal(err)
	}
	write(t, root, ".agents/instructions/team.md", "Use concise commits.\n")

	a, err := NewForTarget(root, "copilot", false)
	if err != nil {
		t.Fatal(err)
	}
	if err := a.Generate(false); err != nil {
		t.Fatalf("Generate() error = %v, want nil when manifest compatibility file is missing", err)
	}
}

func removeSchemaRequiredKey(root, filename, key string) error {
	data, err := os.ReadFile(filepath.Join(root, ".agents", "schema", "v0.0.1", filename))
	if err != nil {
		return err
	}
	var doc map[string]any
	if err := json.Unmarshal(data, &doc); err != nil {
		return err
	}
	rawRequired, ok := doc["required"].([]any)
	if !ok {
		return errors.New("schema missing required key")
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
		return err
	}
	return os.WriteFile(filepath.Join(root, ".agents", "schema", "v0.0.1", filename), updated, 0o644)
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

func mutateMappingsStatusTarget(root, relativePath, rootKey, category, target, value string) error {
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

func removeMappingsStatusCategory(root, relativePath, rootKey, category string) error {
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
	if _, ok := statusRaw[category]; !ok {
		return errors.New("mappings category status missing")
	}
	delete(statusRaw, category)
	updated, err := yaml.Marshal(mappings)
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(root, relativePath), updated, 0o644)
}

func removeMappingsTargetFromStatus(root, relativePath, rootKey, target string) error {
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
	for _, rawCategory := range statusRaw {
		categoryStatuses, ok := rawCategory.(map[string]any)
		if !ok {
			continue
		}
		delete(categoryStatuses, target)
	}
	updated, err := yaml.Marshal(mappings)
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(root, relativePath), updated, 0o644)
}

func mutateManifestValue(root, relativePath, key string, value any) error {
	data, err := os.ReadFile(filepath.Join(root, relativePath))
	if err != nil {
		return err
	}
	var manifest map[string]any
	if err := json.Unmarshal(data, &manifest); err != nil {
		return err
	}
	manifest[key] = value
	updated, err := json.Marshal(manifest)
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
