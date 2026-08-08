package schema

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestGenerateVendorGuideForTargetFilter(t *testing.T) {
	root := t.TempDir()
	seedGuideArtifactsForTests(t, root)

	guide, err := GenerateVendorGuide(root, []string{"codex"})
	if err != nil {
		t.Fatalf("GenerateVendorGuide() error = %v", err)
	}
	if len(guide.Targets) != 1 {
		t.Fatalf("GenerateVendorGuide() targets = %d, want 1", len(guide.Targets))
	}
	if got := guide.Targets[0].ID; got != "codex" {
		t.Fatalf("GenerateVendorGuide() target = %q, want codex", got)
	}
	if len(guide.Categories) == 0 {
		t.Fatal("GenerateVendorGuide() categories empty")
	}
	if guide.Contract.CanonicalRoot != ".agents" {
		t.Fatalf("GenerateVendorGuide() canonical root = %q, want .agents", guide.Contract.CanonicalRoot)
	}
}

func TestGenerateVendorGuideRejectsUnknownTarget(t *testing.T) {
	root := t.TempDir()
	seedGuideArtifactsForTests(t, root)

	if _, err := GenerateVendorGuide(root, []string{"bogus"}); err == nil {
		t.Fatal("GenerateVendorGuide() expected error for unknown target")
	} else if !strings.Contains(err.Error(), "mappings has no status block for target") {
		t.Fatalf("GenerateVendorGuide() error = %q, want unknown-target status error", err)
	}
}

func TestGenerateVendorGuideRequiresTargetDocs(t *testing.T) {
	root := t.TempDir()
	seedGuideArtifactsForTests(t, root)
	if err := removeTargetDocs(root, ".agents/mappings.yaml", ".agents", "codex"); err != nil {
		t.Fatal(err)
	}
	if _, err := GenerateVendorGuide(root, []string{"codex"}); err == nil {
		t.Fatal("GenerateVendorGuide() expected docs validation error")
	} else if !strings.Contains(err.Error(), "mappings target \"codex\" docs must be a non-empty list") {
		t.Fatalf("GenerateVendorGuide() error = %q, want docs missing error", err)
	}
}

func TestRenderVendorGuideMarkdownIncludesTargetAndStatus(t *testing.T) {
	root := t.TempDir()
	seedGuideArtifactsForTests(t, root)

	guide, err := GenerateVendorGuide(root, []string{"copilot"})
	if err != nil {
		t.Fatalf("GenerateVendorGuide() error = %v", err)
	}
	markdown := RenderVendorGuideMarkdown(guide)
	if !strings.Contains(markdown, "# Vendor implementation guide") {
		t.Fatal("RenderVendorGuideMarkdown() missing header")
	}
	if !strings.Contains(markdown, "## Target: GitHub Copilot (copilot)") {
		t.Fatal("RenderVendorGuideMarkdown() missing target heading")
	}
	if !strings.Contains(markdown, "| Category | Status | Notes |") {
		t.Fatal("RenderVendorGuideMarkdown() missing category table header")
	}
	if !strings.Contains(markdown, "| instructions |") {
		t.Fatal("RenderVendorGuideMarkdown() missing category row for instructions")
	}
}

func seedGuideArtifactsForTests(t *testing.T, root string) {
	t.Helper()
	sourceRoot := filepath.Join("..", "..", "..", ".agents")
	if err := copyArtifact(t, filepath.Join(sourceRoot, "schema", "v1", "agents.schema.json"), filepath.Join(root, ".agents", "schema", "v1", "agents.schema.json")); err != nil {
		t.Fatal(err)
	}
	if err := copyArtifact(t, filepath.Join(sourceRoot, "schema", "v1", "mappings.schema.json"), filepath.Join(root, ".agents", "schema", "v1", "mappings.schema.json")); err != nil {
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

func removeTargetDocs(root, relativePath, rootKey, target string) error {
	data, err := os.ReadFile(filepath.Join(root, relativePath))
	if err != nil {
		return err
	}
	var mappings map[string]any
	if err := yaml.Unmarshal(data, &mappings); err != nil {
		return err
	}
	agentsBlock, ok := mappings[rootKey].(map[string]any)
	if !ok {
		return os.ErrInvalid
	}
	targets, ok := agentsBlock["targets"].(map[string]any)
	if !ok {
		return os.ErrInvalid
	}
	targetSpec, ok := targets[target].(map[string]any)
	if !ok {
		return os.ErrInvalid
	}
	delete(targetSpec, "docs")
	updated, err := yaml.Marshal(mappings)
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(root, relativePath), updated, 0o644)
}
