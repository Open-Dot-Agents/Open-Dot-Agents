package dotacore

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Open-Dot-Agents/Open-Dot-Agents/cli/internal/specdata"
	"github.com/Open-Dot-Agents/Open-Dot-Agents/cli/pkg/adapterprotocol"
	jsonschema "github.com/santhosh-tekuri/jsonschema/v5"
)

func TestNormativeTreeFixtures(t *testing.T) {
	root := filepath.Join("..", "..", "..", "conformance", "v1")
	for _, name := range []string{"minimal", "full"} {
		t.Run("valid-"+name, func(t *testing.T) {
			if _, err := Validate(filepath.Join(root, "valid", name)); err != nil {
				t.Fatal(err)
			}
		})
	}
	tests := []struct {
		name string
		code string
	}{{"v0-manifest", "DOTA1001"}, {"bad-rule", "DOTA1201"}, {"profile-cycle", "DOTA1305"}, {"bad-policy", "DOTA1302"}}
	for _, test := range tests {
		t.Run("invalid-"+test.name, func(t *testing.T) {
			_, err := Validate(filepath.Join(root, "invalid", test.name))
			if err == nil || !strings.Contains(err.Error(), test.code) {
				t.Fatalf("error = %v, want %s", err, test.code)
			}
		})
	}
}

func TestNormativeSchemasCompileAndValidateFullFixture(t *testing.T) {
	schemaRoot := filepath.Join("..", "..", "..", "spec", "v1", "schema")
	entries, err := os.ReadDir(schemaRoot)
	if err != nil {
		t.Fatal(err)
	}
	compiler := jsonschema.NewCompiler()
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}
		data, err := os.ReadFile(filepath.Join(schemaRoot, entry.Name()))
		if err != nil {
			t.Fatal(err)
		}
		url := "https://open-dot-agents.github.io/Open-Dot-Agents/spec/v1/schema/" + entry.Name()
		if err := compiler.AddResource(url, bytes.NewReader(data)); err != nil {
			t.Fatal(err)
		}
	}
	fixtureRoot := filepath.Join("..", "..", "..", "conformance", "v1", "valid", "full", ".agents")
	checks := map[string]string{
		"manifest.json":               "manifest.schema.json",
		"hooks/default.json":          "hooks.schema.json",
		"permissions/default.json":    "policy.schema.json",
		"guardrails/default.json":     "guardrails.schema.json",
		"profiles/default.json":       "profiles.schema.json",
		"settings/default.json":       "settings.schema.json",
		"plugins/example/plugin.json": "plugin.schema.json",
		"memories/example.json":       "memory.schema.json",
		"tools/mcp.json":              "mcp.schema.json",
	}
	for fixture, schemaName := range checks {
		t.Run(schemaName, func(t *testing.T) {
			schema, err := compiler.Compile("https://open-dot-agents.github.io/Open-Dot-Agents/spec/v1/schema/" + schemaName)
			if err != nil {
				t.Fatal(err)
			}
			data, err := os.ReadFile(filepath.Join(fixtureRoot, filepath.FromSlash(fixture)))
			if err != nil {
				t.Fatal(err)
			}
			var document any
			if err := json.Unmarshal(data, &document); err != nil {
				t.Fatal(err)
			}
			if err := schema.Validate(document); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestEmbeddedControlSchemasMatchNormativeSources(t *testing.T) {
	entries, err := os.ReadDir(filepath.Join("..", "..", "..", "spec", "v1", "schema"))
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}
		name := entry.Name()
		embedded, err := specdata.Schemas.ReadFile("schema/" + name)
		if err != nil {
			t.Fatal(err)
		}
		source, err := os.ReadFile(filepath.Join("..", "..", "..", "spec", "v1", "schema", name))
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(embedded, source) {
			t.Fatalf("embedded %s differs from normative source", name)
		}
	}
}

func TestApplyPlanRejectsTraversalAndProtectsOwnedChanges(t *testing.T) {
	root := t.TempDir()
	traversal := &adapterprotocol.Plan{Files: []adapterprotocol.File{{Path: "../escape", Encoding: "utf-8", Content: "bad"}}}
	if _, err := ApplyPlan(root, "org.example.test", "export", traversal, ApplyOptions{}); err == nil || !strings.Contains(err.Error(), "DOTA3008") {
		t.Fatalf("traversal error = %v", err)
	}
	plan := &adapterprotocol.Plan{Files: []adapterprotocol.File{{Path: "generated.txt", Encoding: "utf-8", Content: "one\n"}}}
	if _, err := ApplyPlan(root, "org.example.test", "export", plan, ApplyOptions{}); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "generated.txt"), []byte("changed\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := ApplyPlan(root, "org.example.test", "export", plan, ApplyOptions{}); err == nil || !strings.Contains(err.Error(), "DOTA3005") {
		t.Fatalf("modified ownership error = %v", err)
	}
}

func TestApplyAndCheckRejectOutputCollisionsAndInvalidImportNamespaces(t *testing.T) {
	root := t.TempDir()
	collision := &adapterprotocol.Plan{Files: []adapterprotocol.File{
		{Path: "Output.txt", Encoding: "utf-8", Content: "one"},
		{Path: "output.txt", Encoding: "utf-8", Content: "two"},
	}}
	if _, err := ApplyPlan(root, "org.example.test", "export", collision, ApplyOptions{}); err == nil || !strings.Contains(err.Error(), "DOTA3003") {
		t.Fatalf("collision error = %v", err)
	}

	invalidExtension := &adapterprotocol.Plan{Files: []adapterprotocol.File{{
		Path: ".agents/extensions/not-reverse-dns/value.json", Encoding: "utf-8", Content: "{}",
	}}}
	if _, err := ApplyPlan(root, "org.example.test", "import", invalidExtension, ApplyOptions{}); err == nil || !strings.Contains(err.Error(), "DOTA3008") {
		t.Fatalf("extension namespace error = %v", err)
	}

	valid := &adapterprotocol.Plan{Files: []adapterprotocol.File{{Path: "generated.txt", Encoding: "utf-8", Content: "ok"}}}
	if _, err := ApplyPlan(root, "org.example.test", "export", valid, ApplyOptions{}); err != nil {
		t.Fatal(err)
	}
	unsafeCheck := &adapterprotocol.Plan{Files: []adapterprotocol.File{{Path: "../outside", Encoding: "utf-8", Content: "bad"}}}
	if err := CheckPlan(root, "org.example.test", "export", unsafeCheck); err == nil || !strings.Contains(err.Error(), "DOTA3008") {
		t.Fatalf("unsafe check error = %v", err)
	}
}

func TestSnapshotHonorsAdapterByteLimit(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "input.txt"), []byte("12345"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := Snapshot(root, []string{"**"}, 4); err == nil || !strings.Contains(err.Error(), "DOTA3001") {
		t.Fatalf("snapshot error = %v", err)
	}
}

func TestTreeValidationRejectsDuplicateRuleIDsAndFractionalPriorities(t *testing.T) {
	writeManifest := func(t *testing.T, root string) {
		t.Helper()
		if err := os.MkdirAll(filepath.Join(root, ".agents"), 0o755); err != nil {
			t.Fatal(err)
		}
		data := []byte(`{"specVersion":"1.0.0-rc.1","conformance":["core"]}`)
		if err := os.WriteFile(filepath.Join(root, ".agents", "manifest.json"), data, 0o644); err != nil {
			t.Fatal(err)
		}
	}

	t.Run("duplicate rule ids", func(t *testing.T) {
		root := t.TempDir()
		writeManifest(t, root)
		rules := filepath.Join(root, ".agents", "rules")
		if err := os.MkdirAll(rules, 0o755); err != nil {
			t.Fatal(err)
		}
		for _, name := range []string{"one.md", "two.md"} {
			data := []byte("---\nid: duplicate\napplyTo:\n  - '**/*.go'\n---\nRule.\n")
			if err := os.WriteFile(filepath.Join(rules, name), data, 0o644); err != nil {
				t.Fatal(err)
			}
		}
		if _, err := Validate(root); err == nil || !strings.Contains(err.Error(), "DOTA1201") {
			t.Fatalf("duplicate error = %v", err)
		}
	})

	t.Run("fractional instruction priority", func(t *testing.T) {
		root := t.TempDir()
		writeManifest(t, root)
		instructions := filepath.Join(root, ".agents", "instructions")
		if err := os.MkdirAll(instructions, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(instructions, "bad.md"), []byte("---\npriority: 1.5\n---\nInstruction.\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		if _, err := Validate(root); err == nil || !strings.Contains(err.Error(), "DOTA1200") {
			t.Fatalf("priority error = %v", err)
		}
	})
}
