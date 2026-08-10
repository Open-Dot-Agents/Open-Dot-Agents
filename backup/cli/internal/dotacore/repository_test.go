package dotacore

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/Open-Dot-Agents/Open-Dot-Agents/cli/internal/specdata"
)

var markdownLinkPattern = regexp.MustCompile(`\[[^\]]*\]\(([^)]+)\)`)

func TestRepositoryDocumentationLinks(t *testing.T) {
	root := repositoryRoot(t)
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			name := entry.Name()
			if path != root && (name == ".git" || name == ".codegraph" || name == "dist" || name == "bin") {
				return filepath.SkipDir
			}
			return nil
		}
		if filepath.Ext(path) != ".md" {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		for _, match := range markdownLinkPattern.FindAllStringSubmatch(string(data), -1) {
			reference := strings.Trim(strings.TrimSpace(match[1]), "<>")
			if reference == "" || strings.HasPrefix(reference, "#") || strings.Contains(reference, "://") || strings.HasPrefix(reference, "mailto:") {
				continue
			}
			reference = strings.SplitN(reference, "#", 2)[0]
			decoded, err := url.PathUnescape(reference)
			if err != nil {
				return fmt.Errorf("%s: invalid link %q: %w", filepath.ToSlash(path), reference, err)
			}
			target := filepath.Clean(filepath.Join(filepath.Dir(path), filepath.FromSlash(decoded)))
			if _, err := os.Stat(target); err != nil {
				return fmt.Errorf("%s: broken local link %q: %w", filepath.ToSlash(path), reference, err)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestRepositoryVersionMetadata(t *testing.T) {
	root := repositoryRoot(t)
	manifestData, err := os.ReadFile(filepath.Join(root, ".agents", "manifest.json"))
	if err != nil {
		t.Fatal(err)
	}
	var manifest struct {
		SpecVersion string `json:"specVersion"`
	}
	if err := json.Unmarshal(manifestData, &manifest); err != nil {
		t.Fatal(err)
	}
	if manifest.SpecVersion != specdata.SpecVersion {
		t.Fatalf("manifest specVersion = %q, embedded version = %q", manifest.SpecVersion, specdata.SpecVersion)
	}

	expected := map[string]string{
		"README.md":        "current contract is **" + manifest.SpecVersion + "**",
		"COMPATIBILITY.md": "at **" + manifest.SpecVersion + "**",
		"CHANGELOG.md":     "## [v" + manifest.SpecVersion + "]",
	}
	for name, marker := range expected {
		data, err := os.ReadFile(filepath.Join(root, name))
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(data), marker) {
			t.Errorf("%s does not contain version marker %q", name, marker)
		}
	}

	releaseWorkflow, err := os.ReadFile(filepath.Join(root, ".github", "workflows", "release.yml"))
	if err != nil {
		t.Fatal(err)
	}
	for _, marker := range []string{
		"version: v2.17.1",
		"generate-publisher-manifests.sh",
		"inspect-release.sh",
		"dist/publisher-manifests/*.json",
	} {
		if !strings.Contains(string(releaseWorkflow), marker) {
			t.Errorf("release workflow is missing %q", marker)
		}
	}
}

func repositoryRoot(t *testing.T) string {
	t.Helper()
	root, err := filepath.Abs(filepath.Join("..", "..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	return root
}
