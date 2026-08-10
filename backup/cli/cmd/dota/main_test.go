package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Open-Dot-Agents/Open-Dot-Agents/cli/internal/dotacore"
	"github.com/Open-Dot-Agents/Open-Dot-Agents/cli/pkg/adapterprotocol"
)

func TestInitAndValidate(t *testing.T) {
	root := t.TempDir()
	if code, err := run([]string{"init", "--root", root}); code != 0 || err != nil {
		t.Fatalf("init = %d, %v", code, err)
	}
	if code, err := run([]string{"validate", "--root", root}); code != 0 || err != nil {
		t.Fatalf("validate = %d, %v", code, err)
	}
}

func TestAdapterDescriptionMustMatchLockedContract(t *testing.T) {
	locked := dotacore.LockedAdapter{
		ID: "org.example.test", Version: "1.2.3", ProtocolVersion: "1.0",
		Capabilities: []string{"validate", "export"},
	}
	description := adapterprotocol.AdapterDescription{
		ID: "org.example.test", Name: "Test", Version: "1.2.3", ProtocolVersion: "1.0", Target: "test",
		Capabilities: []string{"export", "validate"}, CategoryStatuses: map[string]string{"instructions": "mapped"}, InputPatterns: []string{".agents/**"},
	}
	if err := validateAdapterDescription(locked, description, "exportPlan"); err != nil {
		t.Fatal(err)
	}
	description.Capabilities = []string{"validate"}
	if err := validateAdapterDescription(locked, description, "exportPlan"); err == nil {
		t.Fatal("expected capability mismatch")
	}
	description.Capabilities = []string{"export", "validate"}
	description.ProtocolVersion = "2.0"
	if err := validateAdapterDescription(locked, description, "exportPlan"); err == nil {
		t.Fatal("expected protocol mismatch")
	}
}

func TestValidateRejectsV001(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, ".agents"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, ".agents", "manifest.json"), []byte(`{"format_version":"v0.0.1"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	code, err := run([]string{"validate", "--root", root})
	if code != 1 || err == nil {
		t.Fatalf("validate = %d, %v", code, err)
	}
}

func TestUnknownCommandIsUsageError(t *testing.T) {
	code, err := run([]string{"unknown"})
	if code != 2 || err == nil {
		t.Fatalf("run = %d, %v", code, err)
	}
}
