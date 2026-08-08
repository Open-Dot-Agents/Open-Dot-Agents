package main

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestWriteJSONOutputIsSingleMachineReadableDocument(t *testing.T) {
	var output bytes.Buffer
	results := []commandResult{{Command: "validate", Target: "copilot", Status: "ok"}}

	if err := writeJSONOutput(&output, "validate", results); err != nil {
		t.Fatalf("writeJSONOutput() error = %v", err)
	}

	var decoded commandOutput
	if err := json.Unmarshal(output.Bytes(), &decoded); err != nil {
		t.Fatalf("writeJSONOutput() produced invalid JSON: %v\n%s", err, output.String())
	}
	if decoded.Command != "validate" || len(decoded.Targets) != 1 {
		t.Fatalf("writeJSONOutput() = %#v", decoded)
	}
	if strings.Contains(output.String(), "oda:") {
		t.Fatalf("writeJSONOutput() included human output: %q", output.String())
	}
}

func TestImportRejectsAllTargets(t *testing.T) {
	err := validateCommandTarget("import", "all")
	if err == nil || !strings.Contains(err.Error(), "one explicit source target") {
		t.Fatalf("validateCommandTarget() error = %v", err)
	}
	if err := validateCommandTarget("generate", "all"); err != nil {
		t.Fatalf("generate --target all error = %v", err)
	}
}

func TestImportBacksUpCanonicalDestination(t *testing.T) {
	directory, err := backupDirectory("import", "codex")
	if err != nil {
		t.Fatalf("backupDirectory() error = %v", err)
	}
	if directory != ".agents" {
		t.Fatalf("backupDirectory() = %q, want .agents", directory)
	}

	directory, err = backupDirectory("generate", "codex")
	if err != nil {
		t.Fatalf("backupDirectory() error = %v", err)
	}
	if directory != ".codex" {
		t.Fatalf("backupDirectory() = %q, want .codex", directory)
	}
}
